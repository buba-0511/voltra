package config

import "os"

type Config struct {
	Port          string
	DatabaseDSN   string
	DataDir       string
	AllowedOrigin string
	JWTSecret     string
	OpenAIAPIKey  string
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseDSN:   getEnv("DATABASE_DSN", "postgres://app:app@localhost:5433/energy?sslmode=disable"),
		DataDir:       getEnv("DATA_DIR", "../data"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
