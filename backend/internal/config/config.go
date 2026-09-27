package config

import "os"

// DefaultTenantSlug identifies the single tenant this MVP operates with.
// The schema supports multiple tenants (see internal/db/migrations), but
// there's no tenant-selection UI yet, so both the seed and the login flow
// resolve against this one.
const DefaultTenantSlug = "default"

type Config struct {
	Port          string
	DatabaseDSN   string
	DataDir       string
	AllowedOrigin string
	JWTSecret     string
	OpenAIAPIKey  string
	// CookieSecure marks the auth cookie Secure (HTTPS-only). Off by
	// default so local http:// dev keeps working; set COOKIE_SECURE=true
	// behind HTTPS (Railway, etc).
	CookieSecure bool
}

// Load reads config from the environment. It intentionally has no
// fallback for JWT_SECRET - a silently-applied default secret would mean
// anyone who has seen this (public) source code could forge a valid
// token for any user. Missing it should fail startup, not run insecurely.
func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseDSN:   getEnv("DATABASE_DSN", "postgres://app:app@localhost:5433/energy?sslmode=disable"),
		DataDir:       getEnv("DATA_DIR", "../data"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		CookieSecure:  getEnv("COOKIE_SECURE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
