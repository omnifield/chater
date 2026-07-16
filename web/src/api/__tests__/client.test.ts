import { describe, expect, it, vi } from 'vitest';
import { ApiClient, ApiError } from '../client';
import type { Message } from '../types';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('ApiClient HTTP', () => {
  it('listRooms unwraps {rooms} and sends the bearer token', async () => {
    const fetchImpl = vi.fn(async (_input: string, _init?: RequestInit) =>
      jsonResponse({ rooms: [{ id: 1, type: 'group', title: 't', created_at: 'x' }] }),
    );
    const api = new ApiClient({ getToken: () => 'alice', fetchImpl });

    const rooms = await api.listRooms();

    expect(rooms).toHaveLength(1);
    const init = fetchImpl.mock.calls[0]?.[1];
    const headers = init?.headers as Record<string, string> | undefined;
    expect(headers?.Authorization).toBe('Bearer alice');
  });

  it('sends an ASCII-safe Authorization header for a non-ASCII handle', async () => {
    const fetchImpl = vi.fn(async (_input: string, _init?: RequestInit) =>
      jsonResponse({ rooms: [] }),
    );
    const api = new ApiClient({ getToken: () => 'егор', fetchImpl });

    // Would throw before this resolves if the header were the raw Cyrillic value.
    await api.listRooms();

    const headers = fetchImpl.mock.calls[0]?.[1]?.headers as Record<string, string> | undefined;
    // Exact percent-encoded value — inherently ASCII, a valid header value.
    expect(headers?.Authorization).toBe('Bearer %D0%B5%D0%B3%D0%BE%D1%80');
  });

  it('createRoom posts type and title', async () => {
    const fetchImpl = vi.fn(async (_input: string, _init?: RequestInit) =>
      jsonResponse({ id: 5, type: 'group', title: null, created_at: 'x' }, 201),
    );
    const api = new ApiClient({ getToken: () => 'a', fetchImpl });

    const room = await api.createRoom('group');

    expect(room.id).toBe(5);
    const [url, init] = fetchImpl.mock.calls[0] ?? [];
    expect(url).toContain('/api/chater/rooms');
    expect(JSON.parse(init?.body as string)).toEqual({ type: 'group', title: null });
  });

  it('maps non-ok responses to ApiError with the server message', async () => {
    const fetchImpl = vi.fn(async (_input: string, _init?: RequestInit) =>
      jsonResponse({ error: 'nope' }, 403),
    );
    const api = new ApiClient({ getToken: () => 'a', fetchImpl });

    await expect(api.listRooms()).rejects.toBeInstanceOf(ApiError);
    await expect(api.listRooms()).rejects.toMatchObject({ status: 403, message: 'nope' });
  });

  it('returns undefined on 204', async () => {
    const fetchImpl = vi.fn(
      async (_input: string, _init?: RequestInit) => new Response(null, { status: 204 }),
    );
    const api = new ApiClient({ getToken: () => 'a', fetchImpl });

    await expect(api.addParticipant(1, 2)).resolves.toBeUndefined();
  });

  it('getMessages builds the limit+cursor query', async () => {
    const fetchImpl = vi.fn(async (_input: string, _init?: RequestInit) =>
      jsonResponse({ messages: [], next_cursor: null }),
    );
    const api = new ApiClient({ getToken: () => 'a', fetchImpl });

    await api.getMessages(3, { limit: 2, cursor: 'abc' });

    const url = fetchImpl.mock.calls[0]?.[0];
    expect(url).toContain('/api/chater/rooms/3/messages?limit=2&cursor=abc');
  });

  it('wraps network failures as ApiError(0)', async () => {
    const fetchImpl = vi.fn(async (_input: string, _init?: RequestInit) => {
      throw new Error('boom');
    });
    const api = new ApiClient({ getToken: () => 'a', fetchImpl });

    await expect(api.listRooms()).rejects.toMatchObject({ status: 0 });
  });
});

class FakeSocket {
  static last: FakeSocket | undefined;
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  closed = false;

  constructor(readonly url: string) {
    FakeSocket.last = this;
  }

  close() {
    this.closed = true;
  }
}

describe('ApiClient websocket', () => {
  it('builds the ws url with a token query and delivers message frames', () => {
    const received: Message[] = [];
    let status = '';
    const api = new ApiClient({
      getToken: () => 'alice',
      webSocketImpl: FakeSocket as unknown as new (url: string) => WebSocket,
    });

    const sub = api.openRoomSocket(7, {
      onMessage: (m) => received.push(m),
      onStatus: (s) => {
        status = s;
      },
    });

    const socket = FakeSocket.last;
    expect(socket?.url).toContain('/api/chater/rooms/7/ws?token=alice');

    socket?.onopen?.();
    expect(status).toBe('open');

    socket?.onmessage?.({
      data: JSON.stringify({
        type: 'message',
        message: { id: 1, room_id: 7, author_id: 2, body: 'hi', created_at: 'x' },
      }),
    });
    expect(received).toEqual([{ id: 1, room_id: 7, author_id: 2, body: 'hi', created_at: 'x' }]);

    // non-message / malformed frames are ignored
    socket?.onmessage?.({ data: 'not json' });
    socket?.onmessage?.({ data: JSON.stringify({ type: 'typing' }) });
    expect(received).toHaveLength(1);

    sub.close();
    expect(status).toBe('closed');
    expect(socket?.closed).toBe(true);
  });
});
