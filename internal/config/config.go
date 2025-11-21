package config

import (
	"log"
	"os"
)

type Config struct {
	Port string
}

var AppConfig Config

func LoadConfig() {
	AppConfig = Config{
		Port: getEnv("PORT", "8080"),
	}

	log.Println("Config loaded")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
