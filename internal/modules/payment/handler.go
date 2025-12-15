package payment

import (
	//"ezytix-be/internal/config"
	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	service PaymentService
}

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return &PaymentHandler{service}
}

// ================================================
// 1. WEBHOOK HANDLER (Dipanggil oleh Xendit)
// ================================================
func (h *PaymentHandler) HandleWebhook(c *fiber.Ctx) error {
	// Ambil Token Verifikasi dari Header
	// Xendit mengirim token di header "x-callback-token"
	webhookToken := c.Get("x-callback-token")
	if webhookToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing x-callback-token header",
		})
	}

	var req XenditWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.service.ProcessWebhook(req, webhookToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "webhook processed successfully",
	})
}

// ================================================
// 2. TEST CREATE PAYMENT (Hanya untuk Dev/Testing)
// ================================================
func (h *PaymentHandler) TestCreatePayment(c *fiber.Ctx) error {
	var req CreatePaymentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.service.CreatePayment(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "invoice created",
		"data":    resp,
	})
}