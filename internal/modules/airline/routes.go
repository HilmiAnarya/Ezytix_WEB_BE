package airline

import (
	"ezytix-be/internal/middleware" // Sesuaikan dengan import path projectmu
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AirlineRegisterRoutes(app *fiber.App, db *gorm.DB) {
	repo := NewAirlineRepository(db)
	service := NewAirlineService(repo)
	handler := NewAirlineHandler(service)

	api := app.Group("/api/v1")

	// --- PUBLIC ROUTES ---
	// Endpoint ini bisa diakses publik (atau user login) untuk keperluan searching/filtering
	airlines := api.Group("/airlines")
	airlines.Get("/", handler.GetAllAirlines)    // Untuk Dropdown & Search Filter
	airlines.Get("/:id", handler.GetAirlineByID) // Untuk Detail Info

	// --- ADMIN ROUTES ---
	admin := api.Group("/admin/airlines")

	// Middleware Security
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))

	admin.Post("/", handler.CreateAirline)
	admin.Put("/:id", handler.UpdateAirline)
	admin.Delete("/:id", handler.DeleteAirline)
}