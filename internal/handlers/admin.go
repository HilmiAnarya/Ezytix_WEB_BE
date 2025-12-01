package handlers

import "github.com/gofiber/fiber/v2"

func AdminDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "welcome to admin dashboard",
	})
}
