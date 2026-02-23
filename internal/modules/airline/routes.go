package airline

import (
	"ezytix-be/internal/middleware" 
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AirlineRegisterRoutes(app *fiber.App, db *gorm.DB) {
	repo := NewAirlineRepository(db)
	service := NewAirlineService(repo)
	handler := NewAirlineHandler(service)

	api := app.Group("/api/v1")
	airlines := api.Group("/airlines")
	airlines.Get("/", handler.GetAllAirlines)  
	airlines.Get("/:id", handler.GetAirlineByID)

	admin := api.Group("/admin/airlines")
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RequireRole("admin"))
	admin.Post("/", handler.CreateAirline)
	admin.Put("/:id", handler.UpdateAirline)
	admin.Delete("/:id", handler.DeleteAirline)
}