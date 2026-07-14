package httpapi

import (
	"net/http"
	"strings"

	"github.com/omnifield/chater/internal/store"
)

// --- token-stub identity (NOT real auth) ---
//
// `Authorization: Bearer <token>` where <token> is simply the user's handle.
// The middleware resolves that handle to a user, creating it on first use.
// There is no signature, secret, or expiry — this is a deliberate v0 stub and
// the SINGLE place identity is derived. Swapping in real ecosystem identity
// later means changing only this file: handlers already receive a resolved
// store.User and never see the token.

// authedHandler is a handler that runs with an already-resolved caller.
type authedHandler func(w http.ResponseWriter, r *http.Request, caller store.User)

// withUser wraps a handler with token-stub identity resolution.
func (s *Server) withUser(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, s.logger, http.StatusUnauthorized, "missing or malformed 'Authorization: Bearer <handle>' header")
			return
		}
		caller, err := s.store.GetOrCreateUserByHandle(r.Context(), token)
		if err != nil {
			s.logger.ErrorContext(r.Context(), "resolve caller identity", "err", err)
			writeError(w, s.logger, http.StatusInternalServerError, "identity resolution failed")
			return
		}
		next(w, r, caller)
	}
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(h[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}
