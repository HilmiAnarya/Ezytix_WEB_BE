package server

import (
	"ezytix-be/internal/handlers"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/airport"
	"ezytix-be/internal/modules/flight" // <--- 1. Import Module Flight
	"ezytix-be/internal/middleware"

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
	// ADMIN ROUTES 
	// ================================
	admin := s.App.Group("/api/v1/admin")
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))
}
