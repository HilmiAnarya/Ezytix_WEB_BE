package booking

import (
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/scheduler"
	"ezytix-be/pkg/mail"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func BookingRegisterRoutes(app *fiber.App, db *gorm.DB) {
	bookingRepo := NewBookingRepository(db)
	flightRepo := flight.NewFlightRepository(db)
	authRepo := auth.NewAuthRepository(db)

	flightService := flight.NewFlightService(flightRepo)
	authService := auth.NewAuthService(authRepo, mail.NewMailService()) // [BARU] Tambahkan mail service
	bookingService := NewBookingService(
		bookingRepo, 
		flightService,
		authService,   
	)

	bookingHandler := NewBookingHandler(bookingService)

	api := app.Group("/api/v1")


	bookings := api.Group("/bookings")
	bookings.Use(middleware.JWTMiddleware)
	bookings.Post("/", bookingHandler.CreateOrder)
	bookings.Get("/history", bookingHandler.GetMyBookings)
	bookings.Get("/:order_id/invoice", bookingHandler.DownloadInvoice)
	bookings.Get("/:booking_code/eticket", bookingHandler.DownloadEticket)
	
	scheduler.StartCronJob(bookingService)
}