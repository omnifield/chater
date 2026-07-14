package httpapi

import "sync"

// subBuffer bounds how many undelivered frames a single subscriber may queue
// before the broadcaster starts dropping frames for it (a slow/stuck client must
// never stall the broadcast or the HTTP publisher).
const subBuffer = 32

// subscriber is one websocket connection's inbox. Frames are pre-serialised
// JSON. The channel is never closed — publish() uses a non-blocking send, so a
// gone subscriber simply stops being read and is GC'd after unsubscribe.
type subscriber struct {
	ch chan []byte
}

// hub is an in-process room -> subscribers fan-out. Single-instance only;
// multi-instance pub/sub is a later concern (out of v0 scope).
type hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*subscriber]struct{}
}

func newHub() *hub {
	return &hub{rooms: make(map[int64]map[*subscriber]struct{})}
}

// subscribe registers a new subscriber for a room and returns it.
func (h *hub) subscribe(roomID int64) *subscriber {
	sub := &subscriber{ch: make(chan []byte, subBuffer)}
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.rooms[roomID]
	if subs == nil {
		subs = make(map[*subscriber]struct{})
		h.rooms[roomID] = subs
	}
	subs[sub] = struct{}{}
	return sub
}

// unsubscribe removes a subscriber and drops the room entry once it is empty.
func (h *hub) unsubscribe(roomID int64, sub *subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.rooms[roomID]
	if subs == nil {
		return
	}
	delete(subs, sub)
	if len(subs) == 0 {
		delete(h.rooms, roomID)
	}
}

// publish fans a frame out to every subscriber of a room. The send is
// non-blocking: if a subscriber's buffer is full it drops this frame rather than
// blocking the broadcast (and therefore the HTTP request that triggered it).
func (h *hub) publish(roomID int64, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.rooms[roomID] {
		select {
		case sub.ch <- payload:
		default: // slow/full subscriber — drop this frame, keep broadcasting
		}
	}
}

// subscriberCount reports how many subscribers a room has (used in tests to
// assert lifecycle cleanup).
func (h *hub) subscriberCount(roomID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomID])
}
