// Package store is chater's data-access layer: sqlc-generated types/queries plus
// a thin hand-written wrapper. It returns concrete structs; the consuming
// package (httpapi, later) declares the interface it needs (canon: accept
// interfaces, return structs). Swapping SQLite for Postgres is a driver/DDL
// change behind this boundary, not a rewrite of call sites.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"

	"github.com/omnifield/chater/migrations"
)

// tsLayout is fixed-width UTC ISO-8601 with microseconds and a literal Z, so
// timestamps sort lexicographically in the same order as chronologically — the
// property keyset pagination over created_at relies on.
const tsLayout = "2006-01-02T15:04:05.000000"

func formatTS(t time.Time) string {
	return t.UTC().Format(tsLayout) + "Z"
}

// Store wraps the generated Queries with a clock so timestamps stay injectable
// for tests.
type Store struct {
	q   *Queries
	now func() time.Time
}

// NewStore builds a Store over any database/sql-compatible handle.
func NewStore(db DBTX) *Store {
	return &Store{
		q:   New(db),
		now: func() time.Time { return time.Now().UTC() },
	}
}

// Migrate applies all pending goose migrations against db (idempotent). The
// service calls this on startup; tests call it to build a fresh schema.
func Migrate(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// CreateUser inserts a user with a unique handle.
func (s *Store) CreateUser(ctx context.Context, handle string) (User, error) {
	u, err := s.q.CreateUser(ctx, CreateUserParams{
		Handle:    handle,
		CreatedAt: formatTS(s.now()),
	})
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// CreateRoom inserts a room. roomType is "dialog" or "group"; title may be nil.
func (s *Store) CreateRoom(ctx context.Context, roomType string, title *string) (Room, error) {
	r, err := s.q.CreateRoom(ctx, CreateRoomParams{
		Type:      roomType,
		Title:     nullString(title),
		CreatedAt: formatTS(s.now()),
	})
	if err != nil {
		return Room{}, fmt.Errorf("create room: %w", err)
	}
	return r, nil
}

// AddParticipant adds a user to a room; role may be nil.
func (s *Store) AddParticipant(ctx context.Context, roomID, userID int64, role *string) error {
	if err := s.q.AddParticipant(ctx, AddParticipantParams{
		RoomID:   roomID,
		UserID:   userID,
		Role:     nullString(role),
		JoinedAt: formatTS(s.now()),
	}); err != nil {
		return fmt.Errorf("add participant: %w", err)
	}
	return nil
}

// InsertMessage stores a message and returns it with its assigned id/timestamp.
func (s *Store) InsertMessage(ctx context.Context, roomID, authorID int64, body string) (Message, error) {
	m, err := s.q.InsertMessage(ctx, InsertMessageParams{
		RoomID:    roomID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: formatTS(s.now()),
	})
	if err != nil {
		return Message{}, fmt.Errorf("insert message: %w", err)
	}
	return m, nil
}

// Cursor marks a position in a room's history for keyset pagination. It is the
// (created_at, id) of the last message a caller has seen.
type Cursor struct {
	CreatedAt string
	ID        int64
}

// ListMessagesByRoom returns up to limit messages newest-first. Pass nil cursor
// for the first page, then the cursor built from the last returned message to
// page backwards through older history.
func (s *Store) ListMessagesByRoom(ctx context.Context, roomID int64, cursor *Cursor, limit int32) ([]Message, error) {
	params := ListMessagesByRoomParams{
		RoomID:          roomID,
		CursorCreatedAt: "", // sentinel: no upper bound -> first page
		CursorID:        0,
		PageLimit:       int64(limit),
	}
	if cursor != nil {
		params.CursorCreatedAt = cursor.CreatedAt
		params.CursorID = cursor.ID
	}
	msgs, err := s.q.ListMessagesByRoom(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return msgs, nil
}

// ListRoomsForUser lists every room the user participates in, newest first.
func (s *Store) ListRoomsForUser(ctx context.Context, userID int64) ([]Room, error) {
	rooms, err := s.q.ListRoomsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list rooms for user: %w", err)
	}
	return rooms, nil
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
