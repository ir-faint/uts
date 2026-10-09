package middleware

import (
	"strings"

	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
)

func RequireAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "header Authorization tidak ada atau salah bentuk")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "format token bukan Bearer")
		}

		tokenString := parts[1]

		jwtClaims, err := helper.ParseJWTToken(tokenString, jwtSecret)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "access token tidak valid atau kedaluwarsa")
		}

		c.Locals(helper.LocalsAuthUser, jwtClaims)
		c.Locals("user_id", jwtClaims.UserID)
		c.Locals("email", jwtClaims.Email)
		c.Locals("role", jwtClaims.Role)
		if jwtClaims.StudentID != nil {
			c.Locals("student_id", *jwtClaims.StudentID)
		}

		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}

		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return helper.Fail(c, fiber.StatusForbidden, "Forbidden: Anda tidak memiliki akses ke resource ini")
	}
}

func CheckLoginRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if helper.IsLoginRateLimited(ip) {
			c.Set("Retry-After", "60")
			return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login, coba lagi dalam satu menit")
		}
		return c.Next()
	}
}
