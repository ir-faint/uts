package service

import (
	"fmt"
	"strings"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
)

type EnrollmentService interface {
	CreateEnrollment(c *fiber.Ctx) error
	DeleteEnrollment(c *fiber.Ctx) error
}

type enrollmentService struct {
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
	studentRepo    repository.StudentRepository
}

func NewEnrollmentService(
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
	studentRepo repository.StudentRepository,
) EnrollmentService {
	return &enrollmentService{
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		studentRepo:    studentRepo,
	}
}

func (s *enrollmentService) CreateEnrollment(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, ok := c.Locals("student_id").(int)
	if !ok || studentID <= 0 {
		return helper.Fail(c, fiber.StatusForbidden, "Hanya akun mahasiswa yang dapat mengambil mata kuliah")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "body harus berupa JSON yang valid")
	}

	if errs := ValidateEnrollment(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

	tx, err := s.enrollmentRepo.BeginTx(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memulai transaksi database")
	}
	defer tx.Rollback(ctx)

	// 1. Row Locking FOR UPDATE
	course, err := s.courseRepo.FindByIDForUpdateTx(ctx, tx, req.CourseID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mata kuliah")
	}
	if course == nil {
		return helper.Fail(c, fiber.StatusNotFound, "Mata kuliah tidak ditemukan")
	}

	// 2. Count active enrollments & check quota
	activeEnrollments, err := s.enrollmentRepo.CountActiveEnrollmentsTx(ctx, tx, req.CourseID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghitung kuota terisi")
	}

	sisaKuota := course.Kuota - activeEnrollments
	if sisaKuota <= 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Kuota mata kuliah sudah penuh")
	}

	// 3. Duplicate check
	isDuplicate, err := s.enrollmentRepo.IsAlreadyEnrolledTx(ctx, tx, studentID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa enrollment terdaftar")
	}
	if isDuplicate {
		return helper.Fail(c, fiber.StatusConflict, fmt.Sprintf("Mata kuliah sudah diambil untuk tahun akademik %s", req.TahunAkademik))
	}

	// 4. SKS Limit Check
	student, err := s.studentRepo.FindByID(ctx, studentID)
	if err != nil || student == nil || student.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusNotFound, "Profil mahasiswa tidak ditemukan")
	}

	maxSKS := helper.CalculateBatasSKS(student.IPKTerakhir)
	currentEnrolledSKS, err := s.enrollmentRepo.GetEnrolledSKSInTahunAkademikTx(ctx, tx, studentID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghitung SKS berjalan")
	}

	totalProposedSKS := currentEnrolledSKS + course.SKS
	if totalProposedSKS > maxSKS {
		remainingAvailableSKS := maxSKS - currentEnrolledSKS
		if remainingAvailableSKS < 0 {
			remainingAvailableSKS = 0
		}
		errMsg := fmt.Sprintf(
			"Batas SKS terlampaui. Batas maksimal: %d SKS, SKS diambil: %d SKS, SKS mata kuliah: %d SKS, Sisa SKS tersedia: %d SKS",
			maxSKS, currentEnrolledSKS, course.SKS, remainingAvailableSKS,
		)
		return helper.Fail(c, fiber.StatusUnprocessableEntity, errMsg)
	}

	enrollment := &model.Enrollment{
		StudentID:     studentID,
		CourseID:      req.CourseID,
		TahunAkademik: req.TahunAkademik,
	}

	if err := s.enrollmentRepo.CreateEnrollmentTx(ctx, tx, enrollment); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan enrollment")
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal commit transaksi")
	}

	res := &model.EnrollmentResponse{
		ID:            enrollment.ID,
		StudentID:     enrollment.StudentID,
		CourseID:      enrollment.CourseID,
		TahunAkademik: enrollment.TahunAkademik,
		CreatedAt:     enrollment.CreatedAt,
		Course: &model.CourseResponse{
			ID:        course.ID,
			KodeMK:    course.KodeMK,
			NamaMK:    course.NamaMK,
			SKS:       course.SKS,
			Semester:  course.Semester,
			Kuota:     course.Kuota,
			Terisi:    activeEnrollments + 1,
			SisaKuota: sisaKuota - 1,
		},
	}

	return helper.Created(c, "Enrollment berhasil", res, "")
}

func (s *enrollmentService) DeleteEnrollment(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, ok := c.Locals("student_id").(int)
	if !ok || studentID <= 0 {
		return helper.Fail(c, fiber.StatusForbidden, "Hanya akun mahasiswa yang dapat membatalkan enrollment")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "ID enrollment tidak valid")
	}

	enrollment, err := s.enrollmentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mencari data enrollment")
	}
	if enrollment == nil {
		return helper.Fail(c, fiber.StatusNotFound, "Data enrollment tidak ditemukan")
	}

	if enrollment.StudentID != studentID {
		return helper.Fail(c, fiber.StatusForbidden, "Forbidden: Anda tidak berhak menghapus enrollment mahasiswa lain")
	}

	if err := s.enrollmentRepo.DeleteEnrollment(ctx, id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus enrollment")
	}

	return helper.NoContent(c)
}
