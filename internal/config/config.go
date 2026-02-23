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
	isProd, _ := strconv.ParseBool(getEnv("MIDTRANS_IS_PRODUCTION", "false"))

	AppConfig = Config{
		Port:                 getEnv("PORT", "8080"),
		FrontendURL:          getEnv("FRONTEND_URL", ""),
		MidtransServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey:    getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransIsProduction: isProd,
	}

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