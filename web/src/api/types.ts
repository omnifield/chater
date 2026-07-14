// Wire types — mirror the backend DTOs (internal/httpapi/dto.go).

export type RoomType = 'dialog' | 'group';

export interface User {
  id: number;
  handle: string;
  created_at: string;
}

export interface Room {
  id: number;
  type: RoomType;
  title: string | null;
  created_at: string;
}

export interface Message {
  id: number;
  room_id: number;
  author_id: number;
  body: string;
  created_at: string;
}

export interface MessagesPage {
  messages: Message[];
  next_cursor: string | null;
}

export interface RoomsResponse {
  rooms: Room[];
}

// WebSocket frame envelope: {"type":"message","message":{…}}.
export interface WsEvent {
  type: 'message';
  message?: Message;
}
