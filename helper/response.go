package helper

import (
	"siakad-mini/app/model"

	"github.com/gofiber/fiber/v2"
)

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessList(c *fiber.Ctx, message string, data any, meta *model.Pagination) error {
	return c.Status(fiber.StatusOK).JSON(model.JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	if location != "" {
		c.Set("Location", location)
	}
	return c.Status(fiber.StatusCreated).JSON(model.JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.JSONResponse{
		Success: false,
		Message: message,
	})
}

func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.JSONResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}
