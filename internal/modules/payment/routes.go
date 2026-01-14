package payment

import (
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/booking" // Import modul booking
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func PaymentRegisterRoutes(app *fiber.App, db *gorm.DB) {
	// 1. Setup Repository
	paymentRepo := NewPaymentRepository(db)
	
	// Kita butuh Booking Repo untuk validasi orderID & Amount
	// Inisialisasi on-the-fly di sini agar tidak perlu ubah main.go
	bookingRepo := booking.NewBookingRepository(db) 

	// 2. Setup Service
	// Masukkan bookingRepo sebagai parameter kedua
	service := NewPaymentService(paymentRepo, bookingRepo)
	
	// 3. Setup Handler
	handler := NewPaymentHandler(service)

	// ==========================================
	// 4. DEFINE ROUTES
	// ==========================================
	
	// Group API Payment
	api := app.Group("/api/v1/payments")

	// Endpoint Initiate Payment (Butuh Login)
	// POST /api/v1/payments/initiate
	api.Post("/initiate", middleware.JWTMiddleware, handler.InitiatePayment)

	// Endpoint Webhook (Public - Dipanggil Xendit)
	// POST /api/v1/payments/webhook
	api.Post("/webhook", handler.HandleWebhook)
}