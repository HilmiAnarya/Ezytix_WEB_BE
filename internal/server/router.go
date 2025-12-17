package server

import (
	"ezytix-be/internal/handlers"
	"ezytix-be/internal/middleware"
	
	// --- Import Module ---
	"ezytix-be/internal/modules/airline" // <--- 1. IMPORT INI
	"ezytix-be/internal/modules/airport"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/booking"
	"ezytix-be/internal/modules/flight"
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
	// AIRLINE MODULE (NEW) ✈️
	// ================================
	// <--- 2. REGISTER DISINI
	// Ini penting agar endpoint /airlines dan /admin/airlines bisa diakses
	airline.AirlineRegisterRoutes(s.App, s.DB.GetGORMDB()) 

	// ================================
	// FLIGHT MODULE
	// ================================
	flight.FlightRegisterRoutes(s.App, s.DB.GetGORMDB())

	// ================================
	// Payment MODULE
	// ================================
	payment.PaymentRegisterRoutes(s.App, s.DB.GetGORMDB())

	// ================================
	// BOOKING MODULE
	// ================================
	booking.BookingRegisterRoutes(s.App, s.DB.GetGORMDB())

	// ================================
	// ADMIN ROUTES (Global Admin Group - Optional jika belum dipindah ke modul)
	// ================================
	admin := s.App.Group("/api/v1/admin")
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))
}