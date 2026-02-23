package admin

import (
	"ezytix-be/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
	return &AdminHandler{service}
}

func (h *AdminHandler) GetDashboardStats(c *fiber.Ctx) error {
	claims := c.Locals("user").(*jwt.JWTClaims)
	if claims.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Akses ditolak, hanya untuk admin",
		})
	}

	stats, err := h.service.GetDashboardStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data statistik",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stats,
	})
}