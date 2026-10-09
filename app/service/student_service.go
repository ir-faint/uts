package service

import (
	"math"
	"strconv"
	"strings"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type StudentService interface {
	GetStudents(c *fiber.Ctx) error
	CreateStudent(c *fiber.Ctx) error
	GetStudentByID(c *fiber.Ctx) error
	UpdateStudent(c *fiber.Ctx) error
	DeleteStudent(c *fiber.Ctx) error
}

type studentService struct {
	db          *pgxpool.Pool
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
}

func NewStudentService(db *pgxpool.Pool, userRepo repository.UserRepository, studentRepo repository.StudentRepository) StudentService {
	return &studentService{
		db:          db,
		userRepo:    userRepo,
		studentRepo: studentRepo,
	}
}

func (s *studentService) GetStudents(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	students, total, err := s.studentRepo.FindAll(ctx, q.Page, q.Limit, q.Prodi, q.Angkatan, q.Search, q.Sort)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	lastPage := int(math.Ceil(float64(total) / float64(q.Limit)))
	if lastPage < 1 {
		lastPage = 1
	}

	meta := &model.Pagination{
		CurrentPage: q.Page,
		PerPage:     q.Limit,
		Total:       total,
		LastPage:    lastPage,
	}

	return helper.SuccessList(c, "Daftar mahasiswa berhasil diambil", students, meta)
}

func (s *studentService) CreateStudent(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "body harus berupa JSON yang valid")
	}

	if errs := ValidateStudentCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Email = strings.TrimSpace(req.Email)
	req.Prodi = strings.TrimSpace(req.Prodi)

	existsNIM, err := s.studentRepo.ExistsByNIM(ctx, req.NIM)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa NIM")
	}
	if existsNIM {
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
	}

	existsEmail, err := s.studentRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa email")
	}
	if existsEmail {
		return helper.Fail(c, fiber.StatusConflict, "Email sudah terdaftar")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memulai transaksi database")
	}
	defer tx.Rollback(ctx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NIM), bcrypt.DefaultCost)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal melakukan enkripsi password")
	}

	newUser := &model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "mahasiswa",
	}

	if err := s.userRepo.CreateUserTx(ctx, tx, newUser); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat akun user")
	}

	newStudent := &model.Student{
		UserID:      newUser.ID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	}

	if err := s.studentRepo.CreateStudentTx(ctx, tx, newStudent); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat data mahasiswa")
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal commit transaksi")
	}

	res := &model.StudentResponse{
		ID:          newStudent.ID,
		UserID:      newUser.ID,
		Email:       newUser.Email,
		NIM:         newStudent.NIM,
		Nama:        newStudent.Nama,
		Prodi:       newStudent.Prodi,
		Angkatan:    newStudent.Angkatan,
		IPKTerakhir: newStudent.IPKTerakhir,
	}

	return helper.Created(c, "Mahasiswa berhasil dibuat", res, "/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

func (s *studentService) GetStudentByID(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "ID mahasiswa tidak valid")
	}

	authRole, _ := c.Locals("role").(string)
	authStudentID := 0
	if sID, ok := c.Locals("student_id").(int); ok {
		authStudentID = sID
	}

	if authRole == "mahasiswa" && authStudentID != id {
		return helper.Fail(c, fiber.StatusForbidden, "Forbidden: Anda tidak berhak melihat data mahasiswa lain")
	}

	student, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mencari data mahasiswa")
	}
	if student == nil || student.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	enrolledCourses, totalSKS, err := s.studentRepo.GetEnrolledCoursesByStudentID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar mata kuliah enrolled")
	}

	batasSKS := helper.CalculateBatasSKS(student.IPKTerakhir)

	res := &model.StudentDetailResponse{
		ID:              student.ID,
		UserID:          student.UserID,
		Email:           "",
		NIM:             student.NIM,
		Nama:            student.Nama,
		Prodi:           student.Prodi,
		Angkatan:        student.Angkatan,
		IPKTerakhir:     student.IPKTerakhir,
		BatasSKS:        batasSKS,
		TotalSKSDiambil: totalSKS,
		EnrolledCourses: enrolledCourses,
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", res)
}

func (s *studentService) UpdateStudent(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "ID mahasiswa tidak valid")
	}

	existing, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mencari data mahasiswa")
	}
	if existing == nil || existing.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "body harus berupa JSON yang valid")
	}

	if errs := ValidateStudentUpdate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)

	updated, err := s.studentRepo.UpdateStudent(ctx, id, &req)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui data mahasiswa")
	}

	res := &model.StudentResponse{
		ID:          updated.ID,
		UserID:      updated.UserID,
		NIM:         updated.NIM,
		Nama:        updated.Nama,
		Prodi:       updated.Prodi,
		Angkatan:    updated.Angkatan,
		IPKTerakhir: updated.IPKTerakhir,
	}

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", res)
}

func (s *studentService) DeleteStudent(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "ID mahasiswa tidak valid")
	}

	existing, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mencari data mahasiswa")
	}
	if existing == nil || existing.DeletedAt != nil {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	if err := s.studentRepo.SoftDeleteStudent(ctx, id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus mahasiswa")
	}

	return helper.NoContent(c)
}
