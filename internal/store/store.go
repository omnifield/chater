// Package store is chater's data-access layer: sqlc-generated types/queries plus
// a thin hand-written wrapper. It returns concrete structs; the consuming
// package (httpapi, later) declares the interface it needs (canon: accept
// interfaces, return structs). Swapping SQLite for Postgres is a driver/DDL
// change behind this boundary, not a rewrite of call sites.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"
	msqlite "modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"github.com/omnifield/chater/migrations"
)

// Semantic errors at the store boundary — consumers match these with errors.Is
// instead of depending on database/sql or driver internals.
var (
	// ErrNotFound: a looked-up row does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict: a UNIQUE / PRIMARY KEY constraint was violated
	// (e.g. duplicate handle or duplicate participant).
	ErrConflict = errors.New("conflict")
	// ErrInvalidReference: a FOREIGN KEY constraint was violated
	// (e.g. a participant/room id that does not exist).
	ErrInvalidReference = errors.New("invalid reference")
)

// classifyConstraint maps a driver constraint violation to a semantic error,
// or nil if err is not a recognised constraint violation. Driver-specific
// knowledge stays here, behind the store boundary.
func classifyConstraint(err error) error {
	var se *msqlite.Error
	if !errors.As(err, &se) {
		return nil
	}
	switch se.Code() {
	case sqlitelib.SQLITE_CONSTRAINT_UNIQUE, sqlitelib.SQLITE_CONSTRAINT_PRIMARYKEY:
		return ErrConflict
	case sqlitelib.SQLITE_CONSTRAINT_FOREIGNKEY:
		return ErrInvalidReference
	default:
		return nil
	}
}

// tsLayout is fixed-width UTC ISO-8601 with microseconds and a literal Z, so
// timestamps sort lexicographically in the same order as chronologically — the
// property keyset pagination over created_at relies on.
const tsLayout = "2006-01-02T15:04:05.000000"

func formatTS(t time.Time) string {
	return t.UTC().Format(tsLayout) + "Z"
}

// Store wraps the generated Queries with a clock so timestamps stay injectable
// for tests. It holds the *sql.DB so it can run multi-statement writes in a
// transaction.
type Store struct {
	db  *sql.DB
	q   *Queries
	now func() time.Time
}

// NewStore builds a Store over a database/sql handle.
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:  db,
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
		if c := classifyConstraint(err); c != nil {
			return User{}, c
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetOrCreateUserByHandle resolves a handle to a user, creating it on first
// use. This backs the token-stub identity: the wrapper hides the get→create
// (and the create/get race window) from callers.
func (s *Store) GetOrCreateUserByHandle(ctx context.Context, handle string) (User, error) {
	u, err := s.q.GetUserByHandle(ctx, handle)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return User{}, fmt.Errorf("get user by handle: %w", err)
	}

	created, cerr := s.CreateUser(ctx, handle)
	if cerr == nil {
		return created, nil
	}
	// Lost a create race with a concurrent request for the same handle — the
	// row now exists, so re-read it.
	if u, gerr := s.q.GetUserByHandle(ctx, handle); gerr == nil {
		return u, nil
	}
	return User{}, fmt.Errorf("create user for handle: %w", cerr)
}

// GetRoom returns a room by id, or ErrNotFound.
func (s *Store) GetRoom(ctx context.Context, id int64) (Room, error) {
	r, err := s.q.GetRoom(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Room{}, ErrNotFound
	}
	if err != nil {
		return Room{}, fmt.Errorf("get room: %w", err)
	}
	return r, nil
}

// IsParticipant reports whether user is a member of room.
func (s *Store) IsParticipant(ctx context.Context, roomID, userID int64) (bool, error) {
	ok, err := s.q.RoomParticipantExists(ctx, RoomParticipantExistsParams{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		return false, fmt.Errorf("check participant: %w", err)
	}
	return ok, nil
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

// CreateRoomWithParticipants creates a room and its participants atomically:
// the creator (role "owner") plus any participantIDs (deduped, creator skipped).
// All writes share one transaction, so a bad participant rolls the room back.
func (s *Store) CreateRoomWithParticipants(ctx context.Context, roomType string, title *string, creatorID int64, participantIDs []int64) (Room, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Room{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op after a successful commit

	qtx := s.q.WithTx(tx)
	ts := formatTS(s.now())

	room, err := qtx.CreateRoom(ctx, CreateRoomParams{
		Type:      roomType,
		Title:     nullString(title),
		CreatedAt: ts,
	})
	if err != nil {
		return Room{}, fmt.Errorf("create room: %w", err)
	}

	owner := "owner"
	if err := qtx.AddParticipant(ctx, AddParticipantParams{
		RoomID: room.ID, UserID: creatorID, Role: nullString(&owner), JoinedAt: ts,
	}); err != nil {
		return Room{}, fmt.Errorf("add creator: %w", err)
	}

	seen := map[int64]bool{creatorID: true}
	for _, pid := range participantIDs {
		if seen[pid] {
			continue
		}
		seen[pid] = true
		if err := qtx.AddParticipant(ctx, AddParticipantParams{
			RoomID: room.ID, UserID: pid, Role: sql.NullString{}, JoinedAt: ts,
		}); err != nil {
			if c := classifyConstraint(err); c != nil {
				return Room{}, c
			}
			return Room{}, fmt.Errorf("add participant %d: %w", pid, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Room{}, fmt.Errorf("commit: %w", err)
	}
	return room, nil
}

// AddParticipant adds a user to a room; role may be nil.
func (s *Store) AddParticipant(ctx context.Context, roomID, userID int64, role *string) error {
	if err := s.q.AddParticipant(ctx, AddParticipantParams{
		RoomID:   roomID,
		UserID:   userID,
		Role:     nullString(role),
		JoinedAt: formatTS(s.now()),
	}); err != nil {
		if c := classifyConstraint(err); c != nil {
			return c
		}
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
