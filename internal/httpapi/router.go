// Package httpapi wires chater's HTTP surface using net/http only (Go canon:
// stdlib-first, no framework). Every route lives under the native /chater/
// prefix so the ecosystem gateway can proxy /api/chater/ -> :PORT/chater/
// without rewriting paths.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/omnifield/chater/internal/store"
)

// Prefix is the native URL prefix for all chater routes.
const Prefix = "/chater/"

// Store is the narrow data-access surface httpapi needs. *store.Store satisfies
// it — the consumer declares the interface (canon: accept interfaces), so the
// handlers never depend on the whole store.
type Store interface {
	CreateUser(ctx context.Context, handle string) (store.User, error)
	GetOrCreateUserByHandle(ctx context.Context, handle string) (store.User, error)
	CreateRoomWithParticipants(ctx context.Context, roomType string, title *string, creatorID int64, participantIDs []int64) (store.Room, error)
	ListRoomsForUser(ctx context.Context, userID int64) ([]store.Room, error)
	GetRoom(ctx context.Context, id int64) (store.Room, error)
	AddParticipant(ctx context.Context, roomID, userID int64, role *string) error
	IsParticipant(ctx context.Context, roomID, userID int64) (bool, error)
	InsertMessage(ctx context.Context, roomID, authorID int64, body string) (store.Message, error)
	ListMessagesByRoom(ctx context.Context, roomID int64, cursor *store.Cursor, limit int32) ([]store.Message, error)
}

// Server holds handler dependencies.
type Server struct {
	store  Store
	logger *slog.Logger
	hub    *hub
}

// NewRouter builds the chater HTTP handler over the given store.
func NewRouter(logger *slog.Logger, st Store) http.Handler {
	return newServer(logger, st).routes()
}

func newServer(logger *slog.Logger, st Store) *Server {
	return &Server{store: st, logger: logger, hub: newHub()}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+Prefix+"healthz", healthz(s.logger))

	// Bootstrap: create a user (unauthenticated).
	mux.HandleFunc("POST "+Prefix+"users", s.createUser)

	// Everything else runs behind the token-stub identity middleware.
	mux.HandleFunc("POST "+Prefix+"rooms", s.withUser(s.createRoom))
	mux.HandleFunc("GET "+Prefix+"rooms", s.withUser(s.listRooms))
	mux.HandleFunc("POST "+Prefix+"rooms/{id}/participants", s.withUser(s.addParticipant))
	mux.HandleFunc("POST "+Prefix+"rooms/{id}/messages", s.withUser(s.postMessage))
	mux.HandleFunc("GET "+Prefix+"rooms/{id}/messages", s.withUser(s.getMessages))
	mux.HandleFunc("GET "+Prefix+"rooms/{id}/ws", s.withUser(s.roomWS))

	return mux
}
