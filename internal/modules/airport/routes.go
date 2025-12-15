package airport

import (
	"github.com/gofiber/fiber/v2"
	"ezytix-be/internal/middleware"
	"gorm.io/gorm"
)

func AirportRegisterRoutes(app *fiber.App, db *gorm.DB) {
	repo := NewAirportRepository(db)
	service := NewAirportService(repo)
	handler := NewAirportHandler(service)

	api := app.Group("/api/v1")

	api.Get("/airports", handler.GetAllAirports)
	api.Get("/airports/:id", handler.GetAirportByID)

	admin := api.Group("/admin")

	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))

	admin.Post("/airports", handler.CreateAirport)
	admin.Put("/airports/:id", handler.UpdateAirport)
	admin.Delete("/airports/:id", handler.DeleteAirport)
}
