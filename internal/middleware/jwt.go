package middleware

import (
	"ezytix-be/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func JWTMiddleware(c *fiber.Ctx) error {
	token := c.Cookies("access_token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing access token cookie",
		})
	}

	claims, err := jwt.ValidateAccessToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired access token",
		})
	}

	c.Locals("user", claims)

	return c.Next()
}
