package httpserver

import (
	"net/http"

	"energy-platform/internal/ai"
	"energy-platform/internal/anomalies"
	"energy-platform/internal/auth"
	"energy-platform/internal/config"
	"energy-platform/internal/dashboard"
	sqlcgen "energy-platform/internal/db/sqlc"
	"energy-platform/internal/meters"
)

func New(cfg config.Config, queries *sqlcgen.Queries) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authHandler := auth.NewHandler(queries, cfg.JWTSecret)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	requireAuth := auth.RequireAuth(cfg.JWTSecret)
	mux.Handle("GET /auth/me", requireAuth(http.HandlerFunc(authHandler.Me)))

	metersHandler := meters.NewHandler(queries)
	mux.Handle("GET /meters", requireAuth(http.HandlerFunc(metersHandler.List)))
	mux.Handle("GET /meters/{meterId}", requireAuth(http.HandlerFunc(metersHandler.Get)))
	mux.Handle("GET /meters/{meterId}/readings", requireAuth(http.HandlerFunc(metersHandler.ListReadings)))

	dashboardHandler := dashboard.NewHandler(queries)
	mux.Handle("GET /dashboard/summary", requireAuth(http.HandlerFunc(dashboardHandler.Summary)))

	explainer := ai.NewOpenAIExplainer(cfg.OpenAIAPIKey)
	anomalyService := anomalies.NewService(queries, explainer)
	anomalyHandler := anomalies.NewHandler(anomalyService, queries)
	mux.Handle("POST /ai/analyze", requireAuth(http.HandlerFunc(anomalyHandler.Analyze)))
	mux.Handle("GET /ai/analysis/{id}", requireAuth(http.HandlerFunc(anomalyHandler.GetAnalysisRun)))
	mux.Handle("GET /anomalies", requireAuth(http.HandlerFunc(anomalyHandler.List)))
	mux.Handle("GET /anomalies/{id}", requireAuth(http.HandlerFunc(anomalyHandler.Get)))

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
