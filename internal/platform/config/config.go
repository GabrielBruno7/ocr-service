package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	OCRLanguage  string
	OCRPoolSize  int
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		OCRLanguage: getEnv("OCR_LANGUAGE", "por"),
		OCRPoolSize: getEnvAsInt("OCR_POOL_SIZE", 4),
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
