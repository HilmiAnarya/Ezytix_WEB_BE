package server

import (
	"ezytix-be/internal/handlers"
	//"ezytix-be/internal/middleware"

	// --- Import Module ---
	"ezytix-be/internal/modules/admin"
	"ezytix-be/internal/modules/airline" // <--- 1. IMPORT INI
	"ezytix-be/internal/modules/airport"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/booking"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/modules/payment"

	"github.com/gofiber/contrib/websocket"
)

func (s *FiberServer) RegisterRoutes() {

	s.Get("/", handlers.Home)
	s.Get("/health", handlers.Health)
	s.Get("/ws", websocket.New(handlers.Websocket))
	auth.AuthRegisterRoutes(s.App, s.DB.GetGORMDB())
	airport.AirportRegisterRoutes(s.App, s.DB.GetGORMDB())
	airline.AirlineRegisterRoutes(s.App, s.DB.GetGORMDB())
	flight.FlightRegisterRoutes(s.App, s.DB.GetGORMDB())
	payment.PaymentRegisterRoutes(s.App, s.DB.GetGORMDB())
	booking.BookingRegisterRoutes(s.App, s.DB.GetGORMDB())
	admin.AdminRegisterRoutes(s.App, s.DB.GetGORMDB())

	// admin := s.App.Group("/api/v1/admin")
	// admin.Use(middleware.JWTMiddleware)
	// admin.Use(middleware.RequireRole("admin"))
}