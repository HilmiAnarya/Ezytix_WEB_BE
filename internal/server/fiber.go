package server

import (
	"ezytix-be/internal/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type FiberServer struct {
	*fiber.App
	DB database.Service
}

func New() *FiberServer {
	app := fiber.New(fiber.Config{
		AppName: "ezytix-backend",
	})

	// ============================
	// FIX: CORS UNTUK COOKIES JWT
	// ============================
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true, // <- WAJIB untuk cookie
		ExposeHeaders:    "Set-Cookie",
	}))

	return &FiberServer{
		App: app,
		DB:  database.New(),
	}
}
