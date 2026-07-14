package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/omnifield/chater/internal/store"
)

type createUserRequest struct {
	Handle string `json:"handle"`
}

// createUser bootstraps a user. Unauthenticated by design (chicken-and-egg with
// the token stub, whose token IS a handle).
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if !decodeJSON(w, r, s.logger, &req) {
		return
	}
	handle := strings.TrimSpace(req.Handle)
	if handle == "" {
		writeError(w, s.logger, http.StatusBadRequest, "handle is required")
		return
	}

	u, err := s.store.CreateUser(r.Context(), handle)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(w, s.logger, http.StatusConflict, "handle already exists")
			return
		}
		s.logger.ErrorContext(r.Context(), "create user", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not create user")
		return
	}
	writeJSON(w, s.logger, http.StatusCreated, toUser(u))
}
