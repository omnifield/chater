// Package httpapi wires chater's HTTP surface using net/http only (Go canon:
// stdlib-first, no framework). Every route lives under the native /chater/
// prefix so the ecosystem gateway can proxy /api/chater/ -> :PORT/chater/
// without rewriting paths.
package httpapi

import (
	"log/slog"
	"net/http"
)

// Prefix is the native URL prefix for all chater routes.
const Prefix = "/chater/"

// NewRouter builds the chater HTTP handler.
func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+Prefix+"healthz", healthz(logger))
	return mux
}
