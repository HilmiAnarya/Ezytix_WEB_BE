package payment

import (
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/booking" // Import Booking Module
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PaymentRegisterRoutes adalah entry point yang dipanggil oleh server/router.go
func PaymentRegisterRoutes(app *fiber.App, db *gorm.DB) {
	// 1. Wiring Repository
	paymentRepo := NewPaymentRepository(db)
	
	// Kita butuh Booking Repository untuk Service Payment
	// Asumsi: Modul Booking punya constructor NewBookingRepository yang menerima DB
	bookingRepo := booking.NewBookingRepository(db) 

	// 2. Wiring Service
	// bookingRepo otomatis memenuhi interface BookingServiceContract 
	// (Asal method GetBookingByOrderID & UpdateBookingStatus ada di bookingRepo)
	paymentService := NewPaymentService(paymentRepo, bookingRepo)

	// 3. Wiring Handler
	paymentHandler := NewPaymentHandler(paymentService)

	// 4. Grouping Routes
	api := app.Group("/api/v1/payments")

	// ==========================================
	// ROUTES
	// ==========================================

	// A. Protected Routes (Butuh Login)
	// Menggunakan middleware.JWTMiddleware sesuai standar project kamu
	api.Post("/initiate", middleware.JWTMiddleware, paymentHandler.InitiatePayment)

	// B. Public Routes (Webhook Midtrans)
	// TIDAK BOLEH pakai middleware auth, karena Midtrans yang akses
	api.Post("/webhook", paymentHandler.HandleWebhook)
}