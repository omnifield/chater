package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/omnifield/chater/internal/store"
)

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

// newTestRouter builds a router backed by a real migrated SQLite store in a
// temp file — the handlers are exercised end-to-end against actual SQL.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return NewRouter(logger, store.NewStore(db))
}

// do issues a request and returns the recorder. token=="" omits the auth header.
func do(t *testing.T, router http.Handler, method, target, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, r)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestCreateUser(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "happy path", body: `{"handle":"alice"}`, wantStatus: http.StatusCreated},
		{name: "duplicate handle", body: `{"handle":"alice"}`, wantStatus: http.StatusConflict},
		{name: "missing handle", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "blank handle", body: `{"handle":"   "}`, wantStatus: http.StatusBadRequest},
		{name: "malformed json", body: `{`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, router, http.MethodPost, "/chater/users", "", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestAuthStubRequired(t *testing.T) {
	router := newTestRouter(t)

	// No Authorization header -> 401.
	rec := do(t, router, http.MethodGet, "/chater/rooms", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no-token status = %d, want 401", rec.Code)
	}

	// A Bearer token auto-provisions the user; listing works and starts empty.
	rec = do(t, router, http.MethodGet, "/chater/rooms", "alice", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("with-token status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	got := decodeBody[roomsResponse](t, rec)
	if len(got.Rooms) != 0 {
		t.Fatalf("new user rooms = %d, want 0", len(got.Rooms))
	}
}

func TestCreateRoomAndList(t *testing.T) {
	router := newTestRouter(t)

	// bob exists so alice can add him as a participant on room creation.
	if rec := do(t, router, http.MethodPost, "/chater/users", "", `{"handle":"bob"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create bob: %d", rec.Code)
	}
	bob := decodeBody[userResponse](t, do(t, router, http.MethodPost, "/chater/users", "", `{"handle":"bob2"}`))

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "group with title + participant", body: `{"type":"group","title":"devs","participant_ids":[` + itoa(bob.ID) + `]}`, wantStatus: http.StatusCreated},
		{name: "dialog no title", body: `{"type":"dialog"}`, wantStatus: http.StatusCreated},
		{name: "invalid type", body: `{"type":"channel"}`, wantStatus: http.StatusBadRequest},
		{name: "unknown participant", body: `{"type":"group","participant_ids":[99999]}`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, router, http.MethodPost, "/chater/rooms", "alice", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}

	// alice created two rooms above -> both listed, creator is a member.
	rec := do(t, router, http.MethodGet, "/chater/rooms", "alice", "")
	rooms := decodeBody[roomsResponse](t, rec)
	if len(rooms.Rooms) != 2 {
		t.Fatalf("alice rooms = %d, want 2", len(rooms.Rooms))
	}
}

func TestAddParticipant(t *testing.T) {
	router := newTestRouter(t)
	carol := decodeBody[userResponse](t, do(t, router, http.MethodPost, "/chater/users", "", `{"handle":"carol"}`))

	// alice owns a room.
	room := decodeBody[roomResponse](t, do(t, router, http.MethodPost, "/chater/rooms", "alice", `{"type":"group","title":"t"}`))
	base := "/chater/rooms/" + itoa(room.ID) + "/participants"

	tests := []struct {
		name       string
		token      string
		target     string
		body       string
		wantStatus int
	}{
		{name: "owner adds carol", token: "alice", target: base, body: `{"user_id":` + itoa(carol.ID) + `}`, wantStatus: http.StatusNoContent},
		{name: "duplicate participant", token: "alice", target: base, body: `{"user_id":` + itoa(carol.ID) + `}`, wantStatus: http.StatusConflict},
		{name: "unknown user", token: "alice", target: base, body: `{"user_id":99999}`, wantStatus: http.StatusBadRequest},
		{name: "missing user_id", token: "alice", target: base, body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "non-member forbidden", token: "mallory", target: base, body: `{"user_id":` + itoa(carol.ID) + `}`, wantStatus: http.StatusForbidden},
		{name: "room not found", token: "alice", target: "/chater/rooms/99999/participants", body: `{"user_id":` + itoa(carol.ID) + `}`, wantStatus: http.StatusNotFound},
		{name: "bad room id", token: "alice", target: "/chater/rooms/abc/participants", body: `{"user_id":1}`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, router, http.MethodPost, tt.target, tt.token, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestPostMessageAuthorization(t *testing.T) {
	router := newTestRouter(t)
	room := decodeBody[roomResponse](t, do(t, router, http.MethodPost, "/chater/rooms", "alice", `{"type":"group","title":"t"}`))
	base := "/chater/rooms/" + itoa(room.ID) + "/messages"

	// Participant (owner) can post.
	rec := do(t, router, http.MethodPost, base, "alice", `{"body":"hello"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("owner post = %d, want 201 (%s)", rec.Code, rec.Body.String())
	}
	msg := decodeBody[messageResponse](t, rec)
	if msg.Body != "hello" || msg.RoomID != room.ID {
		t.Fatalf("unexpected message: %+v", msg)
	}

	// Non-participant is forbidden (the 403 enforcement the brief calls out).
	if rec := do(t, router, http.MethodPost, base, "mallory", `{"body":"intrude"}`); rec.Code != http.StatusForbidden {
		t.Fatalf("non-member post = %d, want 403 (%s)", rec.Code, rec.Body.String())
	}

	// Empty body -> 400.
	if rec := do(t, router, http.MethodPost, base, "alice", `{"body":"  "}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty body = %d, want 400", rec.Code)
	}

	// Missing room -> 404.
	if rec := do(t, router, http.MethodPost, "/chater/rooms/99999/messages", "alice", `{"body":"x"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("missing room = %d, want 404", rec.Code)
	}
}

func TestMessageHistoryPaginationOverHTTP(t *testing.T) {
	router := newTestRouter(t)
	room := decodeBody[roomResponse](t, do(t, router, http.MethodPost, "/chater/rooms", "alice", `{"type":"group","title":"t"}`))
	base := "/chater/rooms/" + itoa(room.ID) + "/messages"

	const total = 5
	for i := 0; i < total; i++ {
		body := `{"body":"m` + itoa(int64(i)) + `"}`
		if rec := do(t, router, http.MethodPost, base, "alice", body); rec.Code != http.StatusCreated {
			t.Fatalf("post m%d = %d", i, rec.Code)
		}
	}

	// Non-participant cannot read history.
	if rec := do(t, router, http.MethodGet, base+"?limit=2", "mallory", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("non-member history = %d, want 403", rec.Code)
	}

	// Page through newest-first, 2 at a time, following next_cursor.
	var bodies []string
	target := base + "?limit=2"
	pages := 0
	for {
		rec := do(t, router, http.MethodGet, target, "alice", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("history page = %d (%s)", rec.Code, rec.Body.String())
		}
		page := decodeBody[messagesResponse](t, rec)
		pages++
		for _, m := range page.Messages {
			bodies = append(bodies, m.Body)
		}
		if page.NextCursor == nil {
			break
		}
		target = base + "?limit=2&cursor=" + *page.NextCursor
		if pages > 10 {
			t.Fatal("pagination did not terminate")
		}
	}

	want := []string{"m4", "m3", "m2", "m1", "m0"}
	if strings.Join(bodies, ",") != strings.Join(want, ",") {
		t.Fatalf("history order = %v, want %v", bodies, want)
	}

	// Bad cursor / limit -> 400.
	if rec := do(t, router, http.MethodGet, base+"?cursor=!!bad!!", "alice", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad cursor = %d, want 400", rec.Code)
	}
	if rec := do(t, router, http.MethodGet, base+"?limit=999", "alice", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad limit = %d, want 400", rec.Code)
	}
}
