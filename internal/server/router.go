package server

import (
	"ezytix-be/internal/handlers"
	//"ezytix-be/internal/middleware"
	"ezytix-be/internal/modules/auth"

	"github.com/gofiber/contrib/websocket"
)

func (s *FiberServer) RegisterRoutes() {
	s.Get("/", handlers.Home)
	s.Get("/health", handlers.Health)
	s.Get("/ws", websocket.New(handlers.Websocket))

	// auth module
	auth.AuthRegisterRoutes(s.App, s.DB.GetGORMDB())

	// admin := s.App.Group("/api/v1/admin")

	// // Middleware autentikasi dulu
	// admin.Use(middleware.JWTMiddleware)

	// // Middleware role authorization
	// admin.Use(middleware.RequireRole("admin"))

	// // Endpoint dummy dashboard
	// //admin.Get("/dashboard", handlers.AdminDashboard)
}
