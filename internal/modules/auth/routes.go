package auth

import (
	"ezytix-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AuthRegisterRoutes(app *fiber.App, db *gorm.DB) {
	h := NewAuthHandler(db)

	auth := app.Group("/api/v1/auth")

	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)

	// Protected
	authProtected := auth.Group("/")
	authProtected.Use(middleware.JWTMiddleware)

	authProtected.Get("/me", h.Me)
	authProtected.Post("/logout", h.Logout)
}
