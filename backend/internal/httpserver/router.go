package httpserver

import (
	"context"
	"net/http"
	"time"

	"energy-platform/internal/ai"
	"energy-platform/internal/anomalies"
	"energy-platform/internal/auth"
	"energy-platform/internal/config"
	"energy-platform/internal/dashboard"
	sqlcgen "energy-platform/internal/db/sqlc"
	"energy-platform/internal/meters"
)

// defaultTimeout bounds every ordinary request - without it, a slow or
// hung Postgres query blocks the handler (and holds a pool connection)
// indefinitely instead of failing cleanly.
const defaultTimeout = 10 * time.Second

// analysisTimeout is longer because /ai/analyze legitimately reads and
// scores every meter plus one LLM call per detected anomaly - kept under
// the frontend's own 120s timeout for this endpoint so the backend times
// out first with a clean error instead of the client just giving up.
const analysisTimeout = 90 * time.Second

func New(cfg config.Config, queries *sqlcgen.Queries) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authHandler := auth.NewHandler(queries, cfg.JWTSecret, cfg.CookieSecure)
	mux.Handle("POST /auth/login", withTimeout(defaultTimeout)(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	requireAuth := auth.RequireAuth(cfg.JWTSecret)
	mux.Handle("GET /auth/me", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(authHandler.Me))))

	metersHandler := meters.NewHandler(queries)
	mux.Handle("GET /meters", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(metersHandler.List))))
	mux.Handle("GET /meters/daily", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(metersHandler.AllMetersDailyConsumption))))
	mux.Handle("GET /meters/{meterId}", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(metersHandler.Get))))
	mux.Handle("GET /meters/{meterId}/readings", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(metersHandler.ListReadings))))
	mux.Handle("GET /meters/{meterId}/daily", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(metersHandler.DailyConsumption))))

	dashboardHandler := dashboard.NewHandler(queries)
	mux.Handle("GET /dashboard/summary", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(dashboardHandler.Summary))))

	explainer := ai.NewOpenAIExplainer(cfg.OpenAIAPIKey)
	anomalyService := anomalies.NewService(queries, explainer)
	anomalyHandler := anomalies.NewHandler(anomalyService, queries)
	mux.Handle("POST /ai/analyze", requireAuth(withTimeout(analysisTimeout)(http.HandlerFunc(anomalyHandler.Analyze))))
	mux.Handle("GET /ai/analysis/{id}", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(anomalyHandler.GetAnalysisRun))))
	mux.Handle("GET /anomalies", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(anomalyHandler.List))))
	mux.Handle("GET /anomalies/{id}", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(anomalyHandler.Get))))
	mux.Handle("PATCH /anomalies/{id}", requireAuth(withTimeout(defaultTimeout)(http.HandlerFunc(anomalyHandler.UpdateStatus))))

	return withCORS(cfg.AllowedOrigin, mux)
}

func withTimeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
