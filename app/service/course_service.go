package service

import (
	"strconv"
	"strings"

	"siakad-mini/app/repository"
	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
)

type CourseService interface {
	GetCourses(c *fiber.Ctx) error
}

type courseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(courseRepo repository.CourseRepository) CourseService {
	return &courseService{
		courseRepo: courseRepo,
	}
}

func (s *courseService) GetCourses(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	semester, _ := strconv.Atoi(c.Query("semester", "0"))
	search := strings.TrimSpace(c.Query("search", ""))
	availableStr := c.Query("available", "false")
	available := availableStr == "true" || availableStr == "1"

	courses, err := s.courseRepo.FindAll(ctx, semester, search, available)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar mata kuliah")
	}

	return helper.Success(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", courses)
}
