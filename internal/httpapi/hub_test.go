package httpapi

import "testing"

func TestHubPublishFanOut(t *testing.T) {
	h := newHub()
	const room int64 = 7

	a := h.subscribe(room)
	b := h.subscribe(room)
	if got := h.subscriberCount(room); got != 2 {
		t.Fatalf("subscriberCount = %d, want 2", got)
	}

	h.publish(room, []byte("hello"))

	for name, sub := range map[string]*subscriber{"a": a, "b": b} {
		select {
		case got := <-sub.ch:
			if string(got) != "hello" {
				t.Fatalf("%s got %q, want hello", name, got)
			}
		default:
			t.Fatalf("%s received nothing", name)
		}
	}
}

func TestHubPublishToOtherRoomIsolated(t *testing.T) {
	h := newHub()
	sub := h.subscribe(1)

	h.publish(2, []byte("nope")) // different room

	select {
	case got := <-sub.ch:
		t.Fatalf("subscriber in room 1 unexpectedly got %q", got)
	default:
	}
}

func TestHubUnsubscribeCleansUp(t *testing.T) {
	h := newHub()
	sub := h.subscribe(1)
	h.unsubscribe(1, sub)

	if got := h.subscriberCount(1); got != 0 {
		t.Fatalf("subscriberCount after unsubscribe = %d, want 0", got)
	}
	// room entry itself is dropped once empty
	h.mu.RLock()
	_, present := h.rooms[1]
	h.mu.RUnlock()
	if present {
		t.Fatal("empty room entry should have been removed")
	}
}

func TestHubPublishDropsWhenBufferFull(t *testing.T) {
	h := newHub()
	sub := h.subscribe(1)

	// Overfill: publishing more than subBuffer must not block or panic.
	for i := 0; i < subBuffer+10; i++ {
		h.publish(1, []byte("x"))
	}
	if got := len(sub.ch); got != subBuffer {
		t.Fatalf("buffered = %d, want %d (excess dropped)", got, subBuffer)
	}
}
