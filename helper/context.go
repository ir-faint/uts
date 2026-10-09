package helper

import (
	"siakad-mini/app/model"

	"github.com/gofiber/fiber/v2"
)

const LocalsAuthUser = "authUser"

func CurrentUser(c *fiber.Ctx) (model.JWTClaims, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.JWTClaims)
	return user, ok
}
