package booking

import (
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/modules/payment"
	"ezytix-be/internal/scheduler"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func BookingRegisterRoutes(app *fiber.App, db *gorm.DB) {
	bookingRepo := NewBookingRepository(db)
	paymentRepo := payment.NewPaymentRepository(db)
	flightRepo := flight.NewFlightRepository(db)
	authRepo := auth.NewAuthRepository(db)

	paymentService := payment.NewPaymentService(paymentRepo, bookingRepo)
	flightService := flight.NewFlightService(flightRepo)
	authService := auth.NewAuthService(authRepo)
	bookingService := NewBookingService(
		bookingRepo, 
		flightService,
		paymentService, 
		authService,   
	)

	bookingHandler := NewBookingHandler(bookingService)
	paymentHandler := payment.NewPaymentHandler(paymentService)

	api := app.Group("/api/v1")

	// Route Booking
	bookings := api.Group("/bookings")
	bookings.Use(middleware.JWTMiddleware)
	bookings.Post("/", bookingHandler.CreateOrder)

	// Route Payment Webhook
	payments := api.Group("/payments")
	payments.Post("/webhook", paymentHandler.HandleWebhook)

	scheduler.StartCronJob(bookingService)
}