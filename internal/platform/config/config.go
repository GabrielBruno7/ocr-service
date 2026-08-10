package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	OCRLanguage       string
	OCRPoolSize       int
	DatabaseURL       string
	AMQPURL           string
	UploadDir         string
	ProcessingTimeout time.Duration
}

func Load() Config {
	return Config{
		Port:              getEnv("PORT", "8080"),
		OCRLanguage:       getEnv("OCR_LANGUAGE", "por"),
		OCRPoolSize:       getEnvAsInt("OCR_POOL_SIZE", 4),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://ocr:ocr@postgres:5432/ocr_service?sslmode=disable"),
		AMQPURL:           getEnv("AMQP_URL", "amqp://ocr:ocr@rabbitmq:5672/"),
		UploadDir:         getEnv("UPLOAD_DIR", "/app/uploads"),
		ProcessingTimeout: getEnvAsDuration("PROCESSING_TIMEOUT", 5*time.Minute),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
