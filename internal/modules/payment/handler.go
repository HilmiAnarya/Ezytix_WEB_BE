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

func (h *PaymentHandler) InitiatePayment(c *fiber.Ctx) error {
	var req InitiatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	resp, err := h.service.InitiatePayment(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Payment initiated successfully",
		"data":    resp,
	})
}

func (h *PaymentHandler) HandleWebhook(c *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := h.service.ProcessWebhook(payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Webhook processed successfully",
	})
}

func (h *PaymentHandler) CancelPayment(c *fiber.Ctx) error {
	orderID := c.Params("orderID")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Order ID is required",
		})
	}

	if err := h.service.CancelPayment(orderID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Payment cancelled successfully",
	})
}

func (h *PaymentHandler) GetPaymentStatus(c *fiber.Ctx) error {
	orderID := c.Params("orderID")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Order ID is required",
		})
	}

	resp, err := h.service.GetPaymentByOrderID(orderID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if strings.Contains(err.Error(), "record not found") {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Payment data retrieved successfully",
		"data":    resp,
	})
}