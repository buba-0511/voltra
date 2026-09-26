package httpx

import (
	"encoding/json"
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

// FormatTime renders t in UTC/RFC3339 regardless of the server process's
// local timezone - pgx decodes timestamptz using time.Local, which would
// otherwise make API output depend on where the binary happens to run.
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
