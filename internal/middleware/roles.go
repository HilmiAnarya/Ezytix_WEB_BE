package middleware

import (
	"github.com/gofiber/fiber/v2"
	"ezytix-be/pkg/jwt"
)

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		claims, ok := c.Locals("user").(*jwt.JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		userRole := claims.Role

		for _, role := range allowedRoles {
			if role == userRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden: insufficient permission",
		})
	}
}
