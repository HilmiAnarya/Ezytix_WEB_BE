package payment

import (
	"ezytix-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func PaymentRegisterRoutes(app *fiber.App, db *gorm.DB) {
	repo := NewPaymentRepository(db)

	service := NewPaymentService(repo, nil)
	
	handler := NewPaymentHandler(service)

	// Admin Test Route Only
	admin := app.Group("/api/v1/payments/test")
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))

	admin.Post("/", handler.TestCreatePayment)
}