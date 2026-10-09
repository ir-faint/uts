package config

import (
	"log/slog"

	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
	"siakad-mini/route"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewApp(
	logger *slog.Logger,
	cfg *Config,
	pool *pgxpool.Pool,
	authSvc service.AuthService,
	studentSvc service.StudentService,
	courseSvc service.CourseService,
	enrollmentSvc service.EnrollmentService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "SIAKAD Mini REST API"),
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    1 * 1024 * 1024, // 1 MB limit
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", "*"))

	route.Register(app, route.Dependencies{
		Pool:              pool,
		JWTSecret:         cfg.JWTSecret,
		AuthService:       authSvc,
		StudentService:    studentSvc,
		CourseService:     courseSvc,
		EnrollmentService: enrollmentSvc,
	})

	// Fallback handler for unknown routes
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := "terjadi error pada server"

		if e, ok := err.(*fiber.Error); ok {
			status = e.Code
			message = e.Message
		}

		logger.Error("unhandled_error",
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.String("error", err.Error()),
		)

		return helper.Fail(c, status, message)
	}
}
