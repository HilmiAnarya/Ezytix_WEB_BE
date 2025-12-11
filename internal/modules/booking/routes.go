package booking

import (
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/modules/payment"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func BookingRegisterRoutes(app *fiber.App, db *gorm.DB) {
	// ==========================================
	// 1. DEPENDENCY INJECTION (The Assembly)
	// ==========================================
	
	// A. Siapkan Repository Booking
	bookingRepo := NewBookingRepository(db)

	// B. Siapkan Service Pendukung (Flight, Payment, Auth)
	// Kita perlu menginisialisasi ulang repo & service mereka di sini
	// (Atau idealnya pakai container dependency injection, tapi cara manual ini oke untuk sekarang)
	
	flightRepo := flight.NewFlightRepository(db)
	flightService := flight.NewFlightService(flightRepo)

	paymentRepo := payment.NewPaymentRepository(db)
	paymentService := payment.NewPaymentService(paymentRepo)

	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)

	// C. Siapkan Booking Service (The Orchestrator)
	// Masukkan semua service pendukung ke dalamnya
	bookingService := NewBookingService(
		bookingRepo, 
		flightService, 
		paymentService, 
		authService,
	)

	// D. Siapkan Handler
	handler := NewBookingHandler(bookingService)

	// ==========================================
	// 2. ROUTING
	// ==========================================
	api := app.Group("/api/v1/bookings")

	// Protected Route: Hanya User Login yang bisa booking
	api.Use(middleware.JWTMiddleware)

	// Create Order (Checkout)
	api.Post("/", handler.CreateOrder)
}