package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ezytix-be/internal/config"
	"ezytix-be/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.LoadConfig()

	srv := server.New()
	srv.RegisterRoutes()

	// CHANNEL UNTUK SIGNAL
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// RUN SERVER DALAM GOROUTINE
	go func() {
		log.Printf("Server running on port %s", config.AppConfig.Port)
		if err := srv.Listen(":" + config.AppConfig.Port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// TUNGGU SIGNAL
	<-signalChan
	log.Println("Received shutdown signal...")

	// GRACEFUL SHUTDOWN WITH TIMEOUT
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.ShutdownWithContext(ctx); err != nil {
		log.Println("Forced shutdown:", err)
	}

	// CLOSE DB
	if err := srv.DB.Close(); err != nil {
		log.Println("Error closing DB:", err)
	}

	log.Println("Server exited gracefully.")
}
