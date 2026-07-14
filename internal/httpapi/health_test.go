package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthzRouting(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
	}{
		{name: "healthz ok", method: http.MethodGet, target: "/chater/healthz", wantStatus: http.StatusOK},
		{name: "wrong method", method: http.MethodPost, target: "/chater/healthz", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown route under prefix", method: http.MethodGet, target: "/chater/nope", wantStatus: http.StatusNotFound},
		{name: "no native prefix", method: http.MethodGet, target: "/healthz", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestHealthzResponse(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/chater/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"ok"}`)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
}
