package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/omnifield/chater/internal/store"
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type postMessageRequest struct {
	Body string `json:"body"`
}

// postMessage stores a message authored by the caller. The caller must be a
// participant of the room (403 otherwise).
func (s *Server) postMessage(w http.ResponseWriter, r *http.Request, caller store.User) {
	roomID, ok := pathID(r)
	if !ok {
		writeError(w, s.logger, http.StatusBadRequest, "invalid room id")
		return
	}
	var req postMessageRequest
	if !decodeJSON(w, r, s.logger, &req) {
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		writeError(w, s.logger, http.StatusBadRequest, "body is required")
		return
	}

	if !s.requireParticipant(w, r, roomID, caller.ID) {
		return
	}

	msg, err := s.store.InsertMessage(r.Context(), roomID, caller.ID, body)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "insert message", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not send message")
		return
	}
	writeJSON(w, s.logger, http.StatusCreated, toMessage(msg))
}

type messagesResponse struct {
	Messages   []messageResponse `json:"messages"`
	NextCursor *string           `json:"next_cursor"` // null when there is no next page
}

// getMessages returns one page of room history, newest first. The caller must
// be a participant of the room.
func (s *Server) getMessages(w http.ResponseWriter, r *http.Request, caller store.User) {
	roomID, ok := pathID(r)
	if !ok {
		writeError(w, s.logger, http.StatusBadRequest, "invalid room id")
		return
	}

	limit, ok := parseLimit(r.URL.Query().Get("limit"))
	if !ok {
		writeError(w, s.logger, http.StatusBadRequest, "limit must be between 1 and 100")
		return
	}

	var cursor *store.Cursor
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		c, err := decodeCursor(raw)
		if err != nil {
			writeError(w, s.logger, http.StatusBadRequest, "invalid cursor")
			return
		}
		cursor = &c
	}

	if !s.requireParticipant(w, r, roomID, caller.ID) {
		return
	}

	msgs, err := s.store.ListMessagesByRoom(r.Context(), roomID, cursor, limit)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "list messages", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not load history")
		return
	}

	out := messagesResponse{Messages: make([]messageResponse, 0, len(msgs))}
	for _, m := range msgs {
		out.Messages = append(out.Messages, toMessage(m))
	}
	// A full page implies there may be more; hand back a cursor to continue.
	if len(msgs) == int(limit) {
		last := msgs[len(msgs)-1]
		next := encodeCursor(store.Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
		out.NextCursor = &next
	}
	writeJSON(w, s.logger, http.StatusOK, out)
}

// parseLimit validates the limit query param, applying the default when empty.
func parseLimit(raw string) (int32, bool) {
	if raw == "" {
		return defaultMessageLimit, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxMessageLimit {
		return 0, false
	}
	return int32(n), true
}
