package flight

import (
	"ezytix-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func FlightRegisterRoutes(app *fiber.App, db *gorm.DB) {
	repo := NewFlightRepository(db)
	service := NewFlightService(repo)
	handler := NewFlightHandler(service)

	api := app.Group("/api/v1")

	flights := api.Group("/flights")
	flights.Get("/", handler.GetAllFlights)
	flights.Get("/:id", handler.GetFlightByID)

	admin := api.Group("/admin/flights")
	
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))

	admin.Post("/", handler.CreateFlight)
	
	admin.Put("/:id", handler.UpdateFlight)
	
	admin.Delete("/:id", handler.DeleteFlight)
}