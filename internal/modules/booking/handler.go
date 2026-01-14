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

// ==========================================
// 1. CREATE ORDER HANDLER
// ==========================================
func (h *BookingHandler) CreateOrder(c *fiber.Ctx) error {
	// 1. Safe JWT Extraction (Mencegah Panic)
	userClaims, ok := c.Locals("user").(*jwt.JWTClaims)
	if !ok || userClaims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Invalid token claims",
		})
	}
	userID := userClaims.UserID

	// 2. Parsing Request Body
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
			"error":   err.Error(),
		})
	}

	// 3. Panggil Service
	resp, err := h.service.CreateOrder(userID, req)
	if err != nil {
		// General Error (bisa di-improve dengan mapping error type)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to create order",
			"error":   err.Error(),
		})
	}

	// 4. Return Success
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "order created successfully",
		"data":    resp,
	})
}

// ==========================================
// 2. GET HISTORY HANDLER
// ==========================================
func (h *BookingHandler) GetMyBookings(c *fiber.Ctx) error {
	// 1. Safe JWT Extraction
	userClaims, ok := c.Locals("user").(*jwt.JWTClaims)
	if !ok || userClaims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Invalid token claims",
		})
	}
	userID := userClaims.UserID

	// 2. Panggil Service (Sudah return DTO dengan ExpiryTime yang benar)
	bookings, err := h.service.GetUserBookings(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to fetch booking history",
			"error":   err.Error(),
		})
	}

	// 3. Return Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "successfully fetched booking history",
		"data":    bookings,
	})
}