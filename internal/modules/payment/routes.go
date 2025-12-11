package payment

import (
	"ezytix-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func PaymentRegisterRoutes(app *fiber.App, db *gorm.DB) {
	repo := NewPaymentRepository(db)
	service := NewPaymentService(repo)
	handler := NewPaymentHandler(service)

	api := app.Group("/api/v1/payments")

	// ============================================
	// PUBLIC ROUTES (Webhook Xendit)
	// ============================================
	// Xendit akan menembak ke sini saat pembayaran sukses
	// PENTING: Jangan pasang Middleware JWT di sini!
	api.Post("/webhook", handler.HandleWebhook)

	// ============================================
	// ADMIN/TEST ROUTES (Protected)
	// ============================================
	// Ini hanya untuk testing manual kita via Postman
	// Nanti user biasa create payment lewat flow Booking
	admin := api.Group("/test")
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))

	admin.Post("/", handler.TestCreatePayment)
}