package admin

import (
	"ezytix-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AdminRegisterRoutes(app *fiber.App, db *gorm.DB) {
	repo := NewAdminRepository(db)
	service := NewAdminService(repo)
	handler := NewAdminHandler(service)

	adminGroup := app.Group("/api/v1/admin")

	// Middleware: Wajib Login
	adminGroup.Use(middleware.JWTMiddleware)
	adminGroup.Use(middleware.RequireRole("admin"))

	// Endpoint Statistik
	adminGroup.Get("/dashboard/stats", handler.GetDashboardStats)
}