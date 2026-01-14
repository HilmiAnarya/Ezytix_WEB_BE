package payment

import (
	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	service PaymentService
}

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return &PaymentHandler{service}
}

// ================================================
// 1. INITIATE PAYMENT (Dipanggil Frontend)
// ================================================
func (h *PaymentHandler) InitiatePayment(c *fiber.Ctx) error {
	var req InitiatePaymentRequest

	// Parsing JSON Payload dari Frontend
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
			"error":   err.Error(),
		})
	}

	// Validasi Input (Manual Check)
	if req.OrderID == "" || req.PaymentMethod == "" || req.PaymentType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "missing required fields (order_id, payment_method, payment_type)",
		})
	}

	// Panggil Service
	resp, err := h.service.InitiatePayment(req)
	if err != nil {
		// Kita bisa improve dengan check error type, tapi sementara 500 dulu
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to initiate payment",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "payment initiated successfully",
		"data":    resp,
	})
}

// ================================================
// 2. WEBHOOK HANDLER (Dipanggil Xendit)
// ================================================
func (h *PaymentHandler) HandleWebhook(c *fiber.Ctx) error {
	// Ambil Token Verifikasi dari Header
	webhookToken := c.Get("x-callback-token")
	if webhookToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "missing x-callback-token header",
		})
	}

	// Parsing Payload Dinamis (Map)
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid webhook payload",
			"error":   err.Error(),
		})
	}

	// Proses Webhook di Service
	if err := h.service.ProcessWebhook(payload, webhookToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to process webhook",
			"error":   err.Error(),
		})
	}

	// Return 200 OK agar Xendit tidak retry
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "webhook processed successfully",
	})
}