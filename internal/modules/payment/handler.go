package payment

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	service PaymentService
}

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return &PaymentHandler{service}
}

// POST /api/v1/payments/initiate
func (h *PaymentHandler) InitiatePayment(c *fiber.Ctx) error {
	var req InitiatePaymentRequest
	
	// 1. Parse Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// 2. Call Service
	resp, err := h.service.InitiatePayment(req)
	if err != nil {
		// Mapping HTTP Status Code manual
		status := fiber.StatusBadRequest
		
		if strings.Contains(err.Error(), "booking not found") {
			status = fiber.StatusNotFound
		} else if strings.Contains(err.Error(), "already paid") {
			status = fiber.StatusConflict
		} else if strings.Contains(err.Error(), "expired") {
			status = fiber.StatusGone
		}

		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 3. Success Response (Manual Fiber Map)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Payment initiated successfully",
		"data":    resp,
	})
}

// POST /api/v1/payments/webhook
func (h *PaymentHandler) HandleWebhook(c *fiber.Ctx) error {
	var payload map[string]interface{}

	// 1. Parse Payload JSON dari Midtrans
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid webhook payload",
		})
	}

	// 2. Process di Service
	if err := h.service.ProcessWebhook(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 3. Return 200 OK (Wajib untuk Midtrans)
	return c.SendStatus(fiber.StatusOK)
}