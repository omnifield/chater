import { fireEvent, render, screen, waitFor } from '@solidjs/testing-library';
import { describe, expect, it, vi } from 'vitest';
import type { Message } from '../../api/types';
import { RoomView } from '../RoomView';
import { stubApi } from './testApi';

describe('RoomView', () => {
  it('renders history and sends a message through the api', async () => {
    const history: Message[] = [
      {
        id: 1,
        room_id: 3,
        author_id: 2,
        body: 'earlier',
        created_at: '2026-01-01T00:00:01.000000Z',
      },
    ];
    const sendMessage = vi.fn(async (roomId: number, body: string) => ({
      id: 2,
      room_id: roomId,
      author_id: 1,
      body,
      created_at: '2026-01-01T00:00:02.000000Z',
    }));
    const api = stubApi({
      getMessages: async () => ({ messages: history, next_cursor: null }),
      sendMessage,
    });

    render(() => <RoomView api={api} roomId={3} />);

    expect(await screen.findByText('earlier')).toBeInTheDocument();

    fireEvent.input(screen.getByPlaceholderText('message'), { target: { value: 'hello' } });
    fireEvent.click(screen.getByText('Send'));

    await waitFor(() => expect(sendMessage).toHaveBeenCalledWith(3, 'hello'));
    expect(await screen.findByText('hello')).toBeInTheDocument();
  });

  it('shows the author handle, falling back to #id when absent', async () => {
    const api = stubApi({
      getMessages: async () => ({
        messages: [
          {
            id: 1,
            room_id: 3,
            author_id: 2,
            author_handle: 'alice',
            body: 'named',
            created_at: 'x',
          },
          { id: 2, room_id: 3, author_id: 9, body: 'legacy', created_at: 'y' },
        ],
        next_cursor: null,
      }),
    });

    render(() => <RoomView api={api} roomId={3} />);

    expect(await screen.findByText('alice')).toBeInTheDocument();
    expect(await screen.findByText('#9')).toBeInTheDocument(); // fallback for old payload
  });

  it('appends live messages from the socket', async () => {
    let push: ((m: Message) => void) | undefined;
    const api = stubApi({
      getMessages: async () => ({ messages: [], next_cursor: null }),
      openRoomSocket: (_roomId, handlers) => {
        push = handlers.onMessage;
        return { close: () => {} };
      },
    });

    render(() => <RoomView api={api} roomId={3} />);
    await waitFor(() => expect(push).toBeDefined());

    push?.({
      id: 7,
      room_id: 3,
      author_id: 9,
      body: 'live frame',
      created_at: '2026-01-01T00:00:03.000000Z',
    });

    expect(await screen.findByText('live frame')).toBeInTheDocument();
  });

  it('does not duplicate a sent message echoed by the socket', async () => {
    let push: ((m: Message) => void) | undefined;
    const sent: Message = {
      id: 42,
      room_id: 3,
      author_id: 1,
      body: 'once',
      created_at: '2026-01-01T00:00:04.000000Z',
    };
    const api = stubApi({
      getMessages: async () => ({ messages: [], next_cursor: null }),
      sendMessage: async () => sent,
      openRoomSocket: (_roomId, handlers) => {
        push = handlers.onMessage;
        return { close: () => {} };
      },
    });

    render(() => <RoomView api={api} roomId={3} />);
    await waitFor(() => expect(push).toBeDefined());

    fireEvent.input(screen.getByPlaceholderText('message'), { target: { value: 'once' } });
    fireEvent.click(screen.getByText('Send'));
    await screen.findByText('once');

    // The backend echoes the same message id over the socket.
    push?.(sent);

    await waitFor(() => expect(screen.getAllByText('once')).toHaveLength(1));
  });
});
