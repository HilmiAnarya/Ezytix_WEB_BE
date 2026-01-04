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

func (h *BookingHandler) CreateOrder(c *fiber.Ctx) error {
	claims := c.Locals("user").(*jwt.JWTClaims)
	userID := claims.UserID

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.service.CreateOrder(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "order created successfully",
		"data":    resp,
	})
}

// [FIXED] Handler GetMyBookings (Sesuai Struktur Project)
func (h *BookingHandler) GetMyBookings(c *fiber.Ctx) error {
	// 1. Ambil Claims dari Locals dengan key "user"
	// Middleware sudah memvalidasi token dan menyimpan struct claims di sini
	userClaims, ok := c.Locals("user").(*jwt.JWTClaims)
	
	if !ok || userClaims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Invalid token claims",
		})
	}

	// 2. Ambil UserID (Type uint sesuai struct User)
	userID := userClaims.UserID

	// 3. Panggil Service
	bookings, err := h.service.GetUserBookings(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	// 4. Return Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully fetched booking history",
		"data":    bookings,
	})
}