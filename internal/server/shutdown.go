package server

import (
	"context"
	"log"
	"time"
)

func GracefulShutdown(s *FiberServer) {
	// Hanya melakukan shutdown, tidak menunggu signal
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down server...")

	if err := s.ShutdownWithContext(ctx); err != nil {
		log.Println("Forced shutdown:", err)
	}
}
