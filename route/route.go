package route

import (
	"context"
	"time"

	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWTSecret         string
	AuthService       service.AuthService
	StudentService    service.StudentService
	CourseService     service.CourseService
	EnrollmentService service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// Health Check
	api.Get("/health", healthCheck(deps.Pool))

	jwtSecret := deps.JWTSecret

	// Auth Endpoints
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/login", middleware.CheckLoginRateLimit(), deps.AuthService.Login)
	auth.Get("/me", middleware.RequireAuth(jwtSecret), deps.AuthService.GetMe)

	// Student Endpoints
	students := api.Group("/students", middleware.RequireAuth(jwtSecret))
	students.Get("/", middleware.RequireRole("admin"), deps.StudentService.GetStudents)
	students.Post("/", middleware.RequireRole("admin"), middleware.RequireJSON, deps.StudentService.CreateStudent)
	students.Get("/:id", middleware.RequireRole("admin", "mahasiswa"), deps.StudentService.GetStudentByID)
	students.Put("/:id", middleware.RequireRole("admin"), middleware.RequireJSON, deps.StudentService.UpdateStudent)
	students.Delete("/:id", middleware.RequireRole("admin"), deps.StudentService.DeleteStudent)

	// Course Endpoints
	courses := api.Group("/courses", middleware.RequireAuth(jwtSecret))
	courses.Get("/", middleware.RequireRole("admin", "mahasiswa"), deps.CourseService.GetCourses)

	// Enrollment Endpoints
	enrollments := api.Group("/enrollments", middleware.RequireAuth(jwtSecret))
	enrollments.Post("/", middleware.RequireRole("mahasiswa"), middleware.RequireJSON, deps.EnrollmentService.CreateEnrollment)
	enrollments.Delete("/:id", middleware.RequireRole("mahasiswa"), deps.EnrollmentService.DeleteEnrollment)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
