package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"github.com/omnifield/chater/internal/store"
)

const (
	wsPingInterval = 30 * time.Second
	wsWriteTimeout = 10 * time.Second
	wsPingTimeout  = 10 * time.Second
)

// roomWS upgrades a participant's connection to a websocket and streams live
// room events (receive-only in v0; sending stays on HTTP POST). The membership
// gate runs BEFORE the upgrade, so a non-participant gets a plain 403.
func (s *Server) roomWS(w http.ResponseWriter, r *http.Request, caller store.User) {
	roomID, ok := pathID(r)
	if !ok {
		writeError(w, s.logger, http.StatusBadRequest, "invalid room id")
		return
	}
	// 404 (no room) / 403 (not a participant) are written here, pre-upgrade.
	if !s.requireParticipant(w, r, roomID, caller.ID) {
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Origin/CSRF is not our auth model (the token stub is), and the
		// intended clients are agents/Postman, not browsers. Browser access
		// (can't send an Authorization header on WS) is a known limitation that
		// real identity will address via a subprotocol/query token — not now.
		InsecureSkipVerify: true,
	})
	if err != nil {
		s.logger.WarnContext(r.Context(), "ws accept", "err", err)
		return
	}
	defer func() { _ = c.CloseNow() }()

	sub := s.hub.subscribe(roomID)
	defer s.hub.unsubscribe(roomID, sub)

	// CloseRead drains inbound frames (v0 is receive-only) and, crucially,
	// returns a context that is cancelled when the peer disconnects.
	ctx := c.CloseRead(r.Context())

	ticker := time.NewTicker(wsPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-sub.ch:
			if err := writeFrame(ctx, c, payload); err != nil {
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, wsPingTimeout)
			err := c.Ping(pingCtx)
			cancel()
			if err != nil {
				return // peer unresponsive / gone
			}
		}
	}
}

func writeFrame(ctx context.Context, c *websocket.Conn, payload []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, wsWriteTimeout)
	defer cancel()
	return c.Write(writeCtx, websocket.MessageText, payload)
}
