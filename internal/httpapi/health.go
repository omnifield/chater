package httpapi

import (
	"log/slog"
	"net/http"
)

// healthz reports process liveness: GET /chater/healthz -> 200.
func healthz(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			logger.WarnContext(r.Context(), "write healthz response", "err", err)
		}
	}
}
