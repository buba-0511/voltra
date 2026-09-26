package auth

import (
	"context"
	"net/http"
	"strings"

	"energy-platform/internal/httpx"
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
				httpx.WriteError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
				return
			}
			claims, err := ParseToken(secret, token)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
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
