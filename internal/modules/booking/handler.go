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