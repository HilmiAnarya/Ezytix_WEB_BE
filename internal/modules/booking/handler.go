package booking

import (
	"ezytix-be/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type BookingHandler struct {
	service BookingService
}

func NewBookingHandler(service BookingService) *BookingHandler {
	return &BookingHandler{service}
}

// POST /api/v1/bookings
func (h *BookingHandler) CreateOrder(c *fiber.Ctx) error {
	// 1. Ambil User ID dari JWT (Middleware sudah menjamin ini ada)
	claims := c.Locals("user").(*jwt.JWTClaims)
	userID := claims.UserID

	// 2. Parse Request Body
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// TODO: Validasi struct req menggunakan library validator bisa dipasang di sini

	// 3. Panggil Service
	resp, err := h.service.CreateOrder(userID, req)
	if err != nil {
		// Kita return 500 atau 400 tergantung errornya
		// Untuk simplifikasi, anggap error bisnis & teknis masuk sini
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 4. Return Response (Berisi Payment URL)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "order created successfully",
		"data":    resp,
	})
}