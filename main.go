package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"siakad-mini/app/repository"
	"siakad-mini/app/service"
	"siakad-mini/config"
	"siakad-mini/database"
)

func main() {
	// 1. Konfigurasi dan Logger
	config.LoadEnv()
	cfg := config.LoadConfig()
	logger := config.NewLogger()

	// 2. Database Connection Pool & Auto Migration / Seeding
	pool, err := database.ConnectDB(cfg)
	if err != nil {
		logger.Error("gagal terhubung ke database", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Perakitan Repositori
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	// 4. Perakitan Service
	authSvc := service.NewAuthService(userRepo, studentRepo, cfg.JWTSecret)
	studentSvc := service.NewStudentService(pool, userRepo, studentRepo)
	courseSvc := service.NewCourseService(courseRepo)
	enrollmentSvc := service.NewEnrollmentService(enrollmentRepo, courseRepo, studentRepo)

	// 5. Perakitan Aplikasi Fiber
	app := config.NewApp(logger, cfg, pool, authSvc, studentSvc, courseSvc, enrollmentSvc)
	port := cfg.AppPort

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", "error", err.Error())
			os.Exit(1)
		}
	}()

	logger.Info("server berjalan", "port", port)

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", "error", err.Error())
	}
	logger.Info("server berhenti dengan rapi")
}
