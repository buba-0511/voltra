package httpx

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// WriteServerError logs the real underlying error server-side (nothing
// else in this codebase does - a failure would otherwise vanish into a
// generic client-facing message with no way to diagnose it after the
// fact) and responds with a 500 and the given client-safe message,
// without leaking internals like SQL errors to the response body.
func WriteServerError(w http.ResponseWriter, err error, msg string) {
	log.Printf("%s: %v", msg, err)
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
}

// FormatTime renders t in UTC/RFC3339 regardless of the server process's
// local timezone - pgx decodes timestamptz using time.Local, which would
// otherwise make API output depend on where the binary happens to run.
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
