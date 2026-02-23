package booking

import (
	"ezytix-be/pkg/jwt"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type BookingHandler struct {
	service BookingService
}

func NewBookingHandler(service BookingService) *BookingHandler {
	return &BookingHandler{service}
}

func (h *BookingHandler) CreateOrder(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*jwt.JWTClaims)
	if !ok || userClaims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Invalid token claims",
		})
	}

	userID := userClaims.UserID

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
			"error":   err.Error(),
		})
	}

	resp, err := h.service.CreateOrder(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to create order",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "order created successfully",
		"data":    resp,
	})
}

func (h *BookingHandler) GetMyBookings(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*jwt.JWTClaims)
	if !ok || userClaims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Invalid token claims",
		})
	}

	userID := userClaims.UserID

	bookings, err := h.service.GetUserBookings(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to fetch booking history",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "successfully fetched booking history",
		"data":    bookings,
	})
}

func (h *BookingHandler) DownloadInvoice(c *fiber.Ctx) error {
    orderID := c.Params("order_id")
    if orderID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Order ID is required",
        })
    }

    pdfBytes, err := h.service.DownloadInvoice(c.Context(), orderID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Failed to generate invoice",
            "error":   err.Error(),
        })
    }

    c.Set("Content-Type", "application/pdf")
    c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=Invoice-%s.pdf", orderID))

    return c.Send(pdfBytes)
}

func (h *BookingHandler) DownloadEticket(c *fiber.Ctx) error {
	bookingCode := c.Params("booking_code")
	if bookingCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Booking code is required",
		})
	}

	pdfBytes, err := h.service.DownloadEticket(c.Context(), bookingCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate e-ticket",
			"error":   err.Error(),
		})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=Eticket-%s.pdf", bookingCode))

	return c.Send(pdfBytes)
}