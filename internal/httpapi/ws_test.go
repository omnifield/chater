package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/omnifield/chater/internal/store"
)

// newTestWSServer starts a live httptest server and returns it plus the Server
// (so tests can inspect the hub) and the store-backed handler.
func newTestWSServer(t *testing.T) (*httptest.Server, *Server) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := store.Open(filepath.Join(t.TempDir(), "ws.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	s := newServer(logger, store.NewStore(db))
	srv := httptest.NewServer(s.routes())
	t.Cleanup(srv.Close)
	return srv, s
}

func httpPost(t *testing.T, srv *httptest.Server, path, token, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func createUserAndRoom(t *testing.T, srv *httptest.Server) int64 {
	t.Helper()
	if resp := httpPost(t, srv, "/chater/users", "", `{"handle":"alice"}`); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create alice: %d", resp.StatusCode)
	}
	resp := httpPost(t, srv, "/chater/rooms", "alice", `{"type":"group","title":"t"}`)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create room: %d", resp.StatusCode)
	}
	var room roomResponse
	if err := json.NewDecoder(resp.Body).Decode(&room); err != nil {
		t.Fatalf("decode room: %v", err)
	}
	return room.ID
}

func wsURL(httpURL, path string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http") + path
}

func dialWS(ctx context.Context, srv *httptest.Server, path, token string) (*websocket.Conn, *http.Response, error) {
	opts := &websocket.DialOptions{HTTPHeader: http.Header{}}
	if token != "" {
		opts.HTTPHeader.Set("Authorization", "Bearer "+token)
	}
	return websocket.Dial(ctx, wsURL(srv.URL, path), opts)
}

func readEvent(t *testing.T, ctx context.Context, c *websocket.Conn) wsEvent {
	t.Helper()
	typ, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("ws read: %v", err)
	}
	if typ != websocket.MessageText {
		t.Fatalf("frame type = %v, want text", typ)
	}
	var ev wsEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("unmarshal frame %q: %v", data, err)
	}
	return ev
}

func TestWSLiveDeliveryToTwoSubscribers(t *testing.T) {
	srv, srvObj := newTestWSServer(t)
	roomID := createUserAndRoom(t, srv)
	path := "/chater/rooms/" + itoa(roomID) + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := dialWS(ctx, srv, path, "alice")
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer func() { _ = c1.CloseNow() }()
	c2, _, err := dialWS(ctx, srv, path, "alice")
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	defer func() { _ = c2.CloseNow() }()

	// Both subscriptions must be registered before we publish, else the
	// non-blocking broadcast could race the subscribe. Wait for the hub count.
	waitForSubscribers(t, srvObj.hub, roomID, 2)

	resp := httpPost(t, srv, "/chater/rooms/"+itoa(roomID)+"/messages", "alice", `{"body":"live!"}`)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("post message: %d", resp.StatusCode)
	}

	for name, c := range map[string]*websocket.Conn{"c1": c1, "c2": c2} {
		ev := readEvent(t, ctx, c)
		if ev.Type != "message" || ev.Message == nil || ev.Message.Body != "live!" {
			t.Fatalf("%s got %+v, want message body 'live!'", name, ev)
		}
	}
}

func TestWSNonParticipantRejectedBeforeUpgrade(t *testing.T) {
	srv, _ := newTestWSServer(t)
	roomID := createUserAndRoom(t, srv)
	path := "/chater/rooms/" + itoa(roomID) + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, resp, err := dialWS(ctx, srv, path, "mallory")
	if err == nil {
		_ = c.CloseNow()
		t.Fatal("expected dial to fail for non-participant")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %v, want 403", resp)
	}
}

func TestWSMissingAuthRejected(t *testing.T) {
	srv, _ := newTestWSServer(t)
	roomID := createUserAndRoom(t, srv)
	path := "/chater/rooms/" + itoa(roomID) + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, resp, err := dialWS(ctx, srv, path, "")
	if err == nil {
		_ = c.CloseNow()
		t.Fatal("expected dial to fail without auth")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %v, want 401", resp)
	}
}

func TestWSDisconnectCleansUpSubscription(t *testing.T) {
	srv, srvObj := newTestWSServer(t)
	roomID := createUserAndRoom(t, srv)
	path := "/chater/rooms/" + itoa(roomID) + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := dialWS(ctx, srv, path, "alice")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	waitForSubscribers(t, srvObj.hub, roomID, 1)

	if err := c.Close(websocket.StatusNormalClosure, "bye"); err != nil {
		t.Fatalf("close: %v", err)
	}

	// The server side should detect the close, return from the handler, and
	// unsubscribe — dropping the room back to zero subscribers.
	waitForSubscribers(t, srvObj.hub, roomID, 0)
}

// waitForSubscribers polls the hub until a room reaches want subscribers, or
// fails the test after a short timeout.
func waitForSubscribers(t *testing.T, h *hub, roomID int64, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if got := h.subscriberCount(roomID); got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("subscriberCount(%d) = %d, want %d (timeout)", roomID, h.subscriberCount(roomID), want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
