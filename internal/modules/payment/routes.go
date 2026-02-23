package payment

import (
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/booking"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func PaymentRegisterRoutes(app *fiber.App, db *gorm.DB) {
	paymentRepo := NewPaymentRepository(db)
	bookingRepo := booking.NewBookingRepository(db)
	paymentService := NewPaymentService(paymentRepo, bookingRepo)
	paymentHandler := NewPaymentHandler(paymentService)

	api := app.Group("/api/v1/payments")
	api.Post("/initiate", middleware.JWTMiddleware, paymentHandler.InitiatePayment)
	api.Post("/orders/:orderID/cancel", middleware.JWTMiddleware, paymentHandler.CancelPayment)
	api.Get("/orders/:orderID", middleware.JWTMiddleware, paymentHandler.GetPaymentStatus)
	api.Post("/webhook", paymentHandler.HandleWebhook)
}