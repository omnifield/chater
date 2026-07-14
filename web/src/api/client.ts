import type { Message, MessagesPage, Room, RoomsResponse, RoomType, WsEvent } from './types';

/** ApiError carries the HTTP status so the UI can react (401 -> re-login, etc). */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export type SocketStatus = 'open' | 'reconnecting' | 'closed';

export interface RoomSocketHandlers {
  onMessage: (m: Message) => void;
  onStatus?: (s: SocketStatus) => void;
}

export interface RoomSubscription {
  close: () => void;
}

/**
 * ChatApi is the surface the UI depends on — components take this, never `fetch`
 * directly. ApiClient implements it; tests pass a stub.
 */
export interface ChatApi {
  listRooms(): Promise<Room[]>;
  createRoom(type: RoomType, title?: string): Promise<Room>;
  addParticipant(roomId: number, userId: number, role?: string): Promise<void>;
  getMessages(roomId: number, opts?: { limit?: number; cursor?: string }): Promise<MessagesPage>;
  sendMessage(roomId: number, body: string): Promise<Message>;
  openRoomSocket(roomId: number, handlers: RoomSocketHandlers): RoomSubscription;
}

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;
type WebSocketCtor = new (url: string) => WebSocket;

export interface ApiClientOptions {
  /** API base URL. Empty string = same origin (dev goes through the vite proxy). */
  baseUrl?: string;
  /** Returns the current token (the user's handle), or null when logged out. */
  getToken: () => string | null;
  fetchImpl?: FetchLike;
  webSocketImpl?: WebSocketCtor;
}

const maxBackoffMs = 10_000;

export class ApiClient implements ChatApi {
  private readonly baseUrl: string;
  private readonly getToken: () => string | null;
  private readonly fetchImpl: FetchLike;
  private readonly webSocketImpl?: WebSocketCtor;

  constructor(opts: ApiClientOptions) {
    this.baseUrl = opts.baseUrl ?? '';
    this.getToken = opts.getToken;
    this.fetchImpl = opts.fetchImpl ?? ((input, init) => fetch(input, init));
    this.webSocketImpl = opts.webSocketImpl;
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const headers: Record<string, string> = {};
    const token = this.getToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json';
    }

    let res: Response;
    try {
      res = await this.fetchImpl(this.baseUrl + path, {
        method,
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (cause) {
      throw new ApiError(0, `network error: ${(cause as Error).message}`);
    }

    if (!res.ok) {
      throw new ApiError(res.status, await extractError(res));
    }
    if (res.status === 204) {
      return undefined as T;
    }
    return (await res.json()) as T;
  }

  async listRooms(): Promise<Room[]> {
    const data = await this.request<RoomsResponse>('GET', '/chater/rooms');
    return data.rooms;
  }

  createRoom(type: RoomType, title?: string): Promise<Room> {
    return this.request<Room>('POST', '/chater/rooms', { type, title: title ?? null });
  }

  addParticipant(roomId: number, userId: number, role?: string): Promise<void> {
    return this.request<void>('POST', `/chater/rooms/${roomId}/participants`, {
      user_id: userId,
      role: role ?? null,
    });
  }

  getMessages(roomId: number, opts?: { limit?: number; cursor?: string }): Promise<MessagesPage> {
    const params = new URLSearchParams();
    if (opts?.limit) params.set('limit', String(opts.limit));
    if (opts?.cursor) params.set('cursor', opts.cursor);
    const qs = params.toString();
    const path = `/chater/rooms/${roomId}/messages${qs ? `?${qs}` : ''}`;
    return this.request<MessagesPage>('GET', path);
  }

  sendMessage(roomId: number, body: string): Promise<Message> {
    return this.request<Message>('POST', `/chater/rooms/${roomId}/messages`, { body });
  }

  private wsUrl(roomId: number): string {
    const token = this.getToken() ?? '';
    let proto: string;
    let host: string;
    if (this.baseUrl) {
      const u = new URL(this.baseUrl);
      proto = u.protocol === 'https:' ? 'wss:' : 'ws:';
      host = u.host;
    } else {
      proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
      host = location.host;
    }
    // Token travels as a query param: the browser can't set an Authorization
    // header on a WebSocket, so the dev vite proxy moves ?token= into the header.
    return `${proto}//${host}/chater/rooms/${roomId}/ws?token=${encodeURIComponent(token)}`;
  }

  openRoomSocket(roomId: number, handlers: RoomSocketHandlers): RoomSubscription {
    const Ctor = this.webSocketImpl ?? (globalThis.WebSocket as WebSocketCtor | undefined);
    if (!Ctor) {
      throw new Error('WebSocket is not available in this environment');
    }

    let closed = false;
    let attempt = 0;
    let socket: WebSocket | null = null;
    let timer: ReturnType<typeof setTimeout> | undefined;

    const connect = () => {
      if (closed) return;
      socket = new Ctor(this.wsUrl(roomId));
      socket.onopen = () => {
        attempt = 0;
        handlers.onStatus?.('open');
      };
      socket.onmessage = (ev: MessageEvent) => {
        try {
          const event = JSON.parse(String(ev.data)) as WsEvent;
          if (event.type === 'message' && event.message) {
            handlers.onMessage(event.message);
          }
        } catch {
          // ignore malformed frames
        }
      };
      socket.onclose = () => {
        if (closed) return;
        handlers.onStatus?.('reconnecting');
        const delay = Math.min(500 * 2 ** attempt, maxBackoffMs);
        attempt += 1;
        timer = setTimeout(connect, delay);
      };
      socket.onerror = () => socket?.close();
    };

    connect();

    return {
      close: () => {
        closed = true;
        if (timer) clearTimeout(timer);
        socket?.close();
        handlers.onStatus?.('closed');
      },
    };
  }
}

async function extractError(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string };
    if (data?.error) return data.error;
  } catch {
    // fall through to status text
  }
  return res.statusText || `HTTP ${res.status}`;
}
