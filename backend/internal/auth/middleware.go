package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const claimsContextKey contextKey = "claims"

// RequireAuth validates the Authorization: Bearer <token> header and puts
// the parsed claims on the request context for downstream handlers.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				writeError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
				return
			}
			claims, err := ParseToken(secret, token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsContextKey, claims)))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
