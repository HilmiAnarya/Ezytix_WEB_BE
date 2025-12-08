package flight

import (
	"ezytix-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func FlightRegisterRoutes(app *fiber.App, db *gorm.DB) {
	// 1. Dependency Injection
	// Init Repository -> Service -> Handler
	repo := NewFlightRepository(db)
	service := NewFlightService(repo)
	handler := NewFlightHandler(service)

	// 2. Grouping API Utama
	api := app.Group("/api/v1")

	// ============================================
	// PUBLIC ROUTES (Bisa diakses tanpa login)
	// ============================================
	// User butuh melihat daftar penerbangan dan detailnya
	flights := api.Group("/flights")
	flights.Get("/", handler.GetAllFlights)
	flights.Get("/:id", handler.GetFlightByID)

	// ============================================
	// ADMIN ROUTES (Protected)
	// ============================================
	// Menggunakan Middleware JWT & Role Admin
	admin := api.Group("/admin/flights")
	
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))

	// Create Data
	admin.Post("/", handler.CreateFlight)
	
	// Update Data (Full Replacement Strategy)
	admin.Put("/:id", handler.UpdateFlight)
	
	// Delete Data
	admin.Delete("/:id", handler.DeleteFlight)
}