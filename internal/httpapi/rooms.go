package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/omnifield/chater/internal/store"
)

type createRoomRequest struct {
	Type           string  `json:"type"`
	Title          *string `json:"title"`
	ParticipantIDs []int64 `json:"participant_ids"`
}

type roomsResponse struct {
	Rooms []roomResponse `json:"rooms"`
}

// createRoom creates a room; the caller is added as owner automatically.
func (s *Server) createRoom(w http.ResponseWriter, r *http.Request, caller store.User) {
	var req createRoomRequest
	if !decodeJSON(w, r, s.logger, &req) {
		return
	}
	if req.Type != "dialog" && req.Type != "group" {
		writeError(w, s.logger, http.StatusBadRequest, "type must be 'dialog' or 'group'")
		return
	}
	if req.Title != nil {
		if t := strings.TrimSpace(*req.Title); t == "" {
			req.Title = nil // treat blank title as absent
		} else {
			req.Title = &t
		}
	}

	room, err := s.store.CreateRoomWithParticipants(r.Context(), req.Type, req.Title, caller.ID, req.ParticipantIDs)
	if err != nil {
		if errors.Is(err, store.ErrInvalidReference) {
			writeError(w, s.logger, http.StatusBadRequest, "one or more participant_ids do not exist")
			return
		}
		s.logger.ErrorContext(r.Context(), "create room", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not create room")
		return
	}
	writeJSON(w, s.logger, http.StatusCreated, toRoom(room))
}

// listRooms returns the caller's rooms, newest first.
func (s *Server) listRooms(w http.ResponseWriter, r *http.Request, caller store.User) {
	rooms, err := s.store.ListRoomsForUser(r.Context(), caller.ID)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "list rooms", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not list rooms")
		return
	}
	out := roomsResponse{Rooms: make([]roomResponse, 0, len(rooms))}
	for _, rm := range rooms {
		out.Rooms = append(out.Rooms, toRoom(rm))
	}
	writeJSON(w, s.logger, http.StatusOK, out)
}

type addParticipantRequest struct {
	UserID int64   `json:"user_id"`
	Role   *string `json:"role"`
}

// addParticipant adds a user to a room. Only an existing participant may add
// others.
func (s *Server) addParticipant(w http.ResponseWriter, r *http.Request, caller store.User) {
	roomID, ok := pathID(r)
	if !ok {
		writeError(w, s.logger, http.StatusBadRequest, "invalid room id")
		return
	}
	var req addParticipantRequest
	if !decodeJSON(w, r, s.logger, &req) {
		return
	}
	if req.UserID <= 0 {
		writeError(w, s.logger, http.StatusBadRequest, "user_id is required")
		return
	}

	if !s.requireParticipant(w, r, roomID, caller.ID) {
		return
	}

	err := s.store.AddParticipant(r.Context(), roomID, req.UserID, req.Role)
	switch {
	case errors.Is(err, store.ErrConflict):
		writeError(w, s.logger, http.StatusConflict, "user is already a participant")
	case errors.Is(err, store.ErrInvalidReference):
		writeError(w, s.logger, http.StatusBadRequest, "user_id does not exist")
	case err != nil:
		s.logger.ErrorContext(r.Context(), "add participant", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not add participant")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// requireParticipant enforces that the room exists (404) and the caller is a
// member (403). It writes the error response and returns false on failure.
func (s *Server) requireParticipant(w http.ResponseWriter, r *http.Request, roomID, userID int64) bool {
	if _, err := s.store.GetRoom(r.Context(), roomID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, s.logger, http.StatusNotFound, "room not found")
			return false
		}
		s.logger.ErrorContext(r.Context(), "get room", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not load room")
		return false
	}

	member, err := s.store.IsParticipant(r.Context(), roomID, userID)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "check participant", "err", err)
		writeError(w, s.logger, http.StatusInternalServerError, "could not verify membership")
		return false
	}
	if !member {
		writeError(w, s.logger, http.StatusForbidden, "caller is not a participant of this room")
		return false
	}
	return true
}
