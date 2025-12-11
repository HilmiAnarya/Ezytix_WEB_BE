package config

import (
	"log"
	"os"
)

type Config struct {
	Port                string
	XenditSecretKey     string // [NEW] Untuk Payment Request
	XenditWebhookToken  string // [NEW] Untuk Verifikasi Callback
}

var AppConfig Config

func LoadConfig() {
	AppConfig = Config{
		Port:               getEnv("PORT", "8080"),
		XenditSecretKey:    getEnv("XENDIT_SECRET_KEY", ""),
		XenditWebhookToken: getEnv("XENDIT_WEBHOOK_TOKEN", ""),
	}

	// Validasi Safety: Pastikan key ada saat server start (Opsional tapi recommended)
	if AppConfig.XenditSecretKey == "" {
		log.Println("WARNING: XENDIT_SECRET_KEY is missing in .env")
	}

	log.Println("Config loaded successfully")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}