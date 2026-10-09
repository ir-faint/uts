package service

import (
	"strings"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
)

type AuthService interface {
	Login(c *fiber.Ctx) error
	GetMe(c *fiber.Ctx) error
}

type authService struct {
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
	jwtSecret   string
}

func NewAuthService(userRepo repository.UserRepository, studentRepo repository.StudentRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
		jwtSecret:   jwtSecret,
	}
}

func (s *authService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "body harus berupa JSON yang valid")
	}

	ip := c.IP()
	if helper.IsLoginRateLimited(ip) {
		c.Set("Retry-After", "60")
		return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login, coba lagi dalam satu menit")
	}

	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	user, err := s.userRepo.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data user")
	}

	if user == nil {
		helper.RecordFailedLoginAttempt(ip)
		return helper.Fail(c, fiber.StatusUnauthorized, "email atau password salah")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		helper.RecordFailedLoginAttempt(ip)
		return helper.Fail(c, fiber.StatusUnauthorized, "email atau password salah")
	}

	var studentID *int
	var studentDetail *model.StudentDetail

	if user.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data profil mahasiswa")
		}
		if student == nil || student.DeletedAt != nil {
			helper.RecordFailedLoginAttempt(ip)
			return helper.Fail(c, fiber.StatusUnauthorized, "akun telah dinonaktifkan atau dihapus")
		}
		studentID = &student.ID
		studentDetail = &model.StudentDetail{
			NIM:      student.NIM,
			Nama:     student.Nama,
			Prodi:    student.Prodi,
			Angkatan: student.Angkatan,
			IPK:      &student.IPKTerakhir,
		}
	}

	helper.ResetFailedLoginAttempts(ip)

	token, err := helper.GenerateJWTToken(user, studentID, s.jwtSecret)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token akses")
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", model.LoginResponse{
		Token: token,
		User: model.UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			Role:           user.Role,
			StudentID:      studentID,
			StudentProfile: studentDetail,
		},
	})
}

func (s *authService) GetMe(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	userID, ok := c.Locals("user_id").(int)
	if !ok || userID <= 0 {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil profil user")
	}
	if user == nil {
		return helper.Fail(c, fiber.StatusNotFound, "profil user tidak ditemukan")
	}

	res := &model.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	if strings.ToLower(user.Role) == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data profil mahasiswa")
		}
		if student != nil && student.DeletedAt == nil {
			res.StudentID = &student.ID
			res.StudentProfile = &model.StudentDetail{
				NIM:      student.NIM,
				Nama:     student.Nama,
				Prodi:    student.Prodi,
				Angkatan: student.Angkatan,
				IPK:      &student.IPKTerakhir,
			}
		}
	}

	return helper.Success(c, fiber.StatusOK, "Profil user berhasil diambil", res)
}
