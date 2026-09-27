package auth

import (
	"context"
	"net/http"

	"energy-platform/internal/httpx"
)

type contextKey string

const claimsContextKey contextKey = "claims"

const authCookieName = "token"

// setAuthCookie stores the JWT as an HttpOnly cookie instead of handing
// it back in the JSON body - a token a browser script can read is a
// token an XSS bug can steal. SameSite=None is needed once the cookie is
// Secure (a real cross-site deploy, e.g. separate frontend/backend
// domains); None without Secure is rejected by browsers, and Lax already
// covers the local dev case (same registrable "site", different port).
func setAuthCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteFor(secure),
		MaxAge:   int(tokenTTL.Seconds()),
	})
}

func clearAuthCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteFor(secure),
		MaxAge:   -1,
	})
}

func sameSiteFor(secure bool) http.SameSite {
	if secure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

// RequireAuth validates the HttpOnly auth cookie and puts the parsed
// claims on the request context for downstream handlers.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authCookieName)
			if err != nil || cookie.Value == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
				return
			}
			claims, err := ParseToken(secret, cookie.Value)
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
