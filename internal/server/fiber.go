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

	// GLOBAL MIDDLEWARE (dulunya RegisterGlobalMiddleware)
	app.Use(cors.New())

	return &FiberServer{
		App: app,
		DB:  database.New(),
	}
}
