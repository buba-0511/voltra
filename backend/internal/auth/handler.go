package auth

import (
	"encoding/json"
	"net/http"

	"energy-platform/internal/config"
	sqlcgen "energy-platform/internal/db/sqlc"
	"energy-platform/internal/httpx"
)

type Handler struct {
	Queries      *sqlcgen.Queries
	JWTSecret    string
	CookieSecure bool
}

func NewHandler(q *sqlcgen.Queries, jwtSecret string, cookieSecure bool) *Handler {
	return &Handler{Queries: q, JWTSecret: jwtSecret, CookieSecure: cookieSecure}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	User userPublic `json:"user"`
}

type userPublic struct {
	ID    int32  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.Queries.GetTenantBySlug(r.Context(), config.DefaultTenantSlug)
	if err != nil {
		httpx.WriteServerError(w, err, "tenant lookup failed")
		return
	}

	user, err := h.Queries.GetUserByEmail(r.Context(), sqlcgen.GetUserByEmailParams{
		TenantID: tenant.ID,
		Email:    req.Email,
	})
	if err != nil || !ComparePassword(user.PasswordHash, req.Password) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := GenerateToken(h.JWTSecret, user.ID, user.TenantID, user.Email)
	if err != nil {
		httpx.WriteServerError(w, err, "failed to issue token")
		return
	}

	setAuthCookie(w, token, h.CookieSecure)
	httpx.WriteJSON(w, http.StatusOK, loginResponse{
		User: userPublic{ID: user.ID, Email: user.Email, Name: user.Name},
	})
}

// Logout clears the auth cookie. The JWT itself stays valid until it
// expires (this is a stateless token, not a server-side session), but
// the browser no longer holds it, so the next request has nothing to
// authenticate with - same practical effect for this single-user-per-
// browser app.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w, h.CookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.Queries.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, userPublic{ID: user.ID, Email: user.Email, Name: user.Name})
}
