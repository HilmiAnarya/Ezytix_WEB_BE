package server

import (
	"ezytix-be/internal/handlers"
	"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/airport"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/booking"
	"ezytix-be/internal/modules/flight" // <--- 1. Import Module Flight
	"ezytix-be/internal/modules/payment"

	"github.com/gofiber/contrib/websocket"
)

func (s *FiberServer) RegisterRoutes() {

	// BASIC ROUTES
	s.Get("/", handlers.Home)
	s.Get("/health", handlers.Health)
	s.Get("/ws", websocket.New(handlers.Websocket))

	// ================================
	// AUTH MODULE
	// ================================
	auth.AuthRegisterRoutes(s.App, s.DB.GetGORMDB())

	// ================================
	// AIRPORT MODULE
	// ================================
	airport.AirportRegisterRoutes(s.App, s.DB.GetGORMDB())

	// ================================
	// FLIGHT MODULE (NEW)
	// ================================
	// <--- 2. Register Module Flight disini
	flight.FlightRegisterRoutes(s.App, s.DB.GetGORMDB())

	// ================================
	// Payment MODULE (NEW)
	// ================================
	// <--- 3. Register Module Payment disini
	payment.PaymentRegisterRoutes(s.App, s.DB.GetGORMDB())

	// ================================
    // BOOKING MODULE (THE FINAL PIECE)
    // ================================
    booking.BookingRegisterRoutes(s.App, s.DB.GetGORMDB()) // <--- TAMBAH INI

	// ================================
	// ADMIN ROUTES 
	// ================================
	admin := s.App.Group("/api/v1/admin")
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))
}
