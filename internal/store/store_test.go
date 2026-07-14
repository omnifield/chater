package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

// newTestStore builds a migrated, empty store backed by a temp-file SQLite DB.
// It returns a store whose clock advances one second per call, so timestamps
// are distinct and ordering is deterministic across inserts.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	s := NewStore(db)
	s.now = advancingClock()
	return s
}

// advancingClock returns a deterministic clock that ticks one second per call,
// starting at a fixed instant.
func advancingClock() func() time.Time {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 0
	return func() time.Time {
		t := base.Add(time.Duration(n) * time.Second)
		n++
		return t
	}
}

func ptr(s string) *string { return &s }

func TestMigrateIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("second Migrate (should be no-op): %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "alice")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == 0 || u.Handle != "alice" || u.CreatedAt == "" {
		t.Fatalf("unexpected user: %+v", u)
	}

	// handle is UNIQUE — a duplicate must fail.
	if _, err := s.CreateUser(ctx, "alice"); err == nil {
		t.Fatal("expected duplicate handle to error, got nil")
	}
}

func TestCreateRoom(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		roomType  string
		title     *string
		wantTitle sql.NullString
		wantErr   bool
	}{
		{name: "group with title", roomType: "group", title: ptr("devs"), wantTitle: sql.NullString{String: "devs", Valid: true}},
		{name: "dialog without title", roomType: "dialog", title: nil, wantTitle: sql.NullString{}},
		{name: "invalid type rejected by CHECK", roomType: "channel", title: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := s.CreateRoom(ctx, tt.roomType, tt.title)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateRoom: %v", err)
			}
			if r.Type != tt.roomType || r.Title != tt.wantTitle {
				t.Fatalf("room = %+v, want type=%q title=%+v", r, tt.roomType, tt.wantTitle)
			}
		})
	}
}

func TestForeignKeysEnforced(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// author/room do not exist -> FK violation (pragma foreign_keys is on).
	if _, err := s.InsertMessage(ctx, 999, 999, "hi"); err == nil {
		t.Fatal("expected FK violation inserting message into missing room, got nil")
	}
}

func TestListRoomsForUser(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	alice, err := s.CreateUser(ctx, "alice")
	if err != nil {
		t.Fatalf("CreateUser alice: %v", err)
	}
	bob, err := s.CreateUser(ctx, "bob")
	if err != nil {
		t.Fatalf("CreateUser bob: %v", err)
	}

	r1, _ := s.CreateRoom(ctx, "group", ptr("one"))
	r2, _ := s.CreateRoom(ctx, "dialog", nil)
	// alice is in both rooms; bob only in r1.
	for _, uid := range []int64{alice.ID, bob.ID} {
		if err := s.AddParticipant(ctx, r1.ID, uid, ptr("member")); err != nil {
			t.Fatalf("AddParticipant r1/%d: %v", uid, err)
		}
	}
	if err := s.AddParticipant(ctx, r2.ID, alice.ID, nil); err != nil {
		t.Fatalf("AddParticipant r2/alice: %v", err)
	}

	// composite PK: adding the same (room,user) twice must fail.
	if err := s.AddParticipant(ctx, r1.ID, alice.ID, nil); err == nil {
		t.Fatal("expected duplicate participant to error, got nil")
	}

	rooms, err := s.ListRoomsForUser(ctx, alice.ID)
	if err != nil {
		t.Fatalf("ListRoomsForUser: %v", err)
	}
	if len(rooms) != 2 {
		t.Fatalf("alice rooms = %d, want 2", len(rooms))
	}
	// newest-first: r2 created after r1.
	if rooms[0].ID != r2.ID || rooms[1].ID != r1.ID {
		t.Fatalf("room order = [%d,%d], want [%d,%d]", rooms[0].ID, rooms[1].ID, r2.ID, r1.ID)
	}

	bobRooms, err := s.ListRoomsForUser(ctx, bob.ID)
	if err != nil {
		t.Fatalf("ListRoomsForUser bob: %v", err)
	}
	if len(bobRooms) != 1 || bobRooms[0].ID != r1.ID {
		t.Fatalf("bob rooms = %+v, want just r1", bobRooms)
	}
}

func TestMessageHistoryPagination(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	author, _ := s.CreateUser(ctx, "author")
	room, _ := s.CreateRoom(ctx, "group", ptr("chat"))

	const total = 5
	bodies := []string{"m0", "m1", "m2", "m3", "m4"} // inserted in order, ascending time
	for _, b := range bodies {
		if _, err := s.InsertMessage(ctx, room.ID, author.ID, b); err != nil {
			t.Fatalf("InsertMessage %s: %v", b, err)
		}
	}

	// Page through newest-first, 2 at a time, following the keyset cursor.
	var got []string
	var cursor *Cursor
	pages := 0
	for {
		page, err := s.ListMessagesByRoom(ctx, room.ID, cursor, 2)
		if err != nil {
			t.Fatalf("ListMessagesByRoom: %v", err)
		}
		if len(page) == 0 {
			break
		}
		pages++
		for _, m := range page {
			got = append(got, m.Body)
		}
		last := page[len(page)-1]
		cursor = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		if len(page) < 2 {
			break
		}
	}

	// Expect strict newest-first order, every message exactly once.
	want := []string{"m4", "m3", "m2", "m1", "m0"}
	if len(got) != total {
		t.Fatalf("paged %d messages, want %d (%v)", len(got), total, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("page order = %v, want %v", got, want)
		}
	}
	if pages != 3 { // 2 + 2 + 1
		t.Fatalf("pages = %d, want 3", pages)
	}
}

func TestMessageHistoryKeysetTiebreak(t *testing.T) {
	// Two messages with the SAME timestamp must still paginate deterministically
	// via the id tiebreaker in the keyset predicate.
	s := newTestStore(t)
	s.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) } // frozen clock
	ctx := context.Background()

	author, _ := s.CreateUser(ctx, "author")
	room, _ := s.CreateRoom(ctx, "group", nil)

	first, _ := s.InsertMessage(ctx, room.ID, author.ID, "first")
	second, _ := s.InsertMessage(ctx, room.ID, author.ID, "second")
	if first.CreatedAt != second.CreatedAt {
		t.Fatalf("timestamps not equal: %q vs %q", first.CreatedAt, second.CreatedAt)
	}

	page1, err := s.ListMessagesByRoom(ctx, room.ID, nil, 1)
	if err != nil || len(page1) != 1 || page1[0].Body != "second" {
		t.Fatalf("page1 = %+v (err=%v), want [second]", page1, err)
	}
	cursor := &Cursor{CreatedAt: page1[0].CreatedAt, ID: page1[0].ID}
	page2, err := s.ListMessagesByRoom(ctx, room.ID, cursor, 1)
	if err != nil || len(page2) != 1 || page2[0].Body != "first" {
		t.Fatalf("page2 = %+v (err=%v), want [first]", page2, err)
	}
}
