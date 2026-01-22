package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	FrontendURL          string
	
	// [UPDATED] Midtrans Config
	MidtransServerKey    string
	MidtransClientKey    string
	MidtransIsProduction bool
}

var AppConfig Config

func LoadConfig() {
	// Parsing boolean untuk Production Mode (default false)
	isProd, _ := strconv.ParseBool(getEnv("MIDTRANS_IS_PRODUCTION", "false"))

	AppConfig = Config{
		Port:                 getEnv("PORT", "8080"),
		FrontendURL:          getEnv("FRONTEND_URL", ""),
		
		// Load Key Midtrans dari ENV
		MidtransServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey:    getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransIsProduction: isProd,
	}

	// Validasi Safety
	if AppConfig.MidtransServerKey == "" {
		log.Println("WARNING: MIDTRANS_SERVER_KEY is missing in .env")
	}

	log.Println("Config loaded successfully")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}