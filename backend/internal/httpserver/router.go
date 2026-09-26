package httpserver

import (
	"net/http"

	"energy-platform/internal/auth"
	"energy-platform/internal/config"
	sqlcgen "energy-platform/internal/db/sqlc"
)

func New(cfg config.Config, queries *sqlcgen.Queries) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authHandler := auth.NewHandler(queries, cfg.JWTSecret)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("GET /auth/me", auth.RequireAuth(cfg.JWTSecret)(http.HandlerFunc(authHandler.Me)))

	return withCORS(cfg.AllowedOrigin, mux)
}

func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
