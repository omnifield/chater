package httpapi

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/omnifield/chater/internal/store"
)

// --- token-stub identity (NOT real auth) ---
//
// `Authorization: Bearer <token>` where <token> is the user's handle,
// percent-encoded. Encoding keeps the token ASCII so a non-ASCII handle (e.g.
// Cyrillic) yields a valid HTTP header value — a raw non-ASCII header value is
// rejected by browsers/clients before the request is even sent. A plain ASCII
// handle percent-encodes to itself, so existing `Bearer alice` tokens are
// unaffected. The middleware decodes the token, then resolves (creating on first
// use) the user. No signature, secret, or expiry — a deliberate v0 stub and the
// SINGLE place identity is derived. Real ecosystem identity later changes only
// this file: handlers already receive a resolved store.User.

// handleFromToken decodes a bearer token back to a handle. A percent-encoded
// token round-trips to the original handle; a token with no escapes (or a
// malformed one) is used verbatim, so raw ASCII tokens keep working.
func handleFromToken(token string) string {
	if decoded, err := url.PathUnescape(token); err == nil {
		return decoded
	}
	return token
}

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
		caller, err := s.store.GetOrCreateUserByHandle(r.Context(), handleFromToken(token))
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
