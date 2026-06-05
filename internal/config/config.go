package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string
	BackendURL     string
	UploadDir      string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	port := getEnv("PORT", "8080")
	return &Config{
		Port:           port,
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		BackendURL:     getEnv("BACKEND_URL", "http://localhost:"+port),
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
