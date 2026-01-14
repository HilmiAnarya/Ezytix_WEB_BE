package booking

import (
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/scheduler"
	
	// ❌ HAPUS IMPORT PAYMENT
	// "ezytix-be/internal/modules/payment" 

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func BookingRegisterRoutes(app *fiber.App, db *gorm.DB) {
	// 1. Setup Repository
	bookingRepo := NewBookingRepository(db)
	flightRepo := flight.NewFlightRepository(db)
	authRepo := auth.NewAuthRepository(db)

	// ❌ HAPUS PaymentRepo

	// 2. Setup Service
	flightService := flight.NewFlightService(flightRepo)
	authService := auth.NewAuthService(authRepo)
	
	// ❌ HAPUS PaymentService init
	
	// [REFACTORED] Constructor BookingService Baru
	// Parameter paymentService SUDAH DIHAPUS.
	bookingService := NewBookingService(
		bookingRepo, 
		flightService,
		authService,   
	)

	// 3. Setup Handler
	bookingHandler := NewBookingHandler(bookingService)
	
	// ❌ HAPUS PaymentHandler

	// 4. Register Routes
	api := app.Group("/api/v1")

	// Route Booking
	bookings := api.Group("/bookings")
	bookings.Use(middleware.JWTMiddleware)
	
	bookings.Post("/", bookingHandler.CreateOrder)
	bookings.Get("/history", bookingHandler.GetMyBookings)

	// ❌ ROUTE PAYMENT & WEBHOOK DIHAPUS DARI SINI
	// (Sudah dipindahkan ke internal/modules/payment/routes.go)

	// 5. Start Scheduler
	scheduler.StartCronJob(bookingService)
}