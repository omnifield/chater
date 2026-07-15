package httpapi

import "github.com/omnifield/chater/internal/store"

// Wire DTOs. Handlers map store structs to these so the JSON contract is
// explicit and never leaks database types (e.g. sql.NullString).

type userResponse struct {
	ID        int64  `json:"id"`
	Handle    string `json:"handle"`
	CreatedAt string `json:"created_at"`
}

func toUser(u store.User) userResponse {
	return userResponse{ID: u.ID, Handle: u.Handle, CreatedAt: u.CreatedAt}
}

type roomResponse struct {
	ID        int64   `json:"id"`
	Type      string  `json:"type"`
	Title     *string `json:"title"` // null for titleless rooms (e.g. dialogs)
	CreatedAt string  `json:"created_at"`
}

func toRoom(r store.Room) roomResponse {
	var title *string
	if r.Title.Valid {
		t := r.Title.String
		title = &t
	}
	return roomResponse{ID: r.ID, Type: r.Type, Title: title, CreatedAt: r.CreatedAt}
}

type messageResponse struct {
	ID           int64  `json:"id"`
	RoomID       int64  `json:"room_id"`
	AuthorID     int64  `json:"author_id"`
	AuthorHandle string `json:"author_handle"`
	Body         string `json:"body"`
	CreatedAt    string `json:"created_at"`
}

func toMessage(m store.Message, authorHandle string) messageResponse {
	return messageResponse{
		ID:           m.ID,
		RoomID:       m.RoomID,
		AuthorID:     m.AuthorID,
		AuthorHandle: authorHandle,
		Body:         m.Body,
		CreatedAt:    m.CreatedAt,
	}
}

// wsEvent is the websocket frame envelope. The `type` discriminator leaves room
// for future event kinds (typing, presence, …); v0 emits only "message". The
// payload reuses the HTTP message DTO so there is a single wire shape.
type wsEvent struct {
	Type    string           `json:"type"`
	Message *messageResponse `json:"message,omitempty"`
}

func messageEvent(m store.Message, authorHandle string) wsEvent {
	dto := toMessage(m, authorHandle)
	return wsEvent{Type: "message", Message: &dto}
}
