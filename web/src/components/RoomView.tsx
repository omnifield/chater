import { createEffect, createSignal, For, onCleanup, Show } from 'solid-js';
import type { ChatApi, SocketStatus } from '../api/client';
import type { Message } from '../api/types';
import { errorText } from '../format';

const PAGE = 30;

/**
 * RoomView shows a room's history (with keyset "load older"), sends messages,
 * and subscribes to the live websocket feed. Sent and live messages are merged
 * by id, so the echo of one's own POST does not duplicate.
 */
export function RoomView(props: { api: ChatApi; roomId: number }) {
  const [messages, setMessages] = createSignal<Message[]>([]);
  const [cursor, setCursor] = createSignal<string | null>(null);
  const [status, setStatus] = createSignal<SocketStatus | ''>('');
  const [error, setError] = createSignal<string | null>(null);
  const [draft, setDraft] = createSignal('');
  const [participantId, setParticipantId] = createSignal('');

  const upsert = (incoming: Message[]) => {
    setMessages((prev) => {
      const byId = new Map(prev.map((m) => [m.id, m] as const));
      for (const m of incoming) byId.set(m.id, m);
      return [...byId.values()].sort((a, b) =>
        a.created_at === b.created_at ? a.id - b.id : a.created_at < b.created_at ? -1 : 1,
      );
    });
  };

  // Reload history and (re)subscribe whenever the room changes. onCleanup inside
  // the effect closes the previous socket before opening the next.
  createEffect(() => {
    const roomId = props.roomId;
    setMessages([]);
    setCursor(null);
    setError(null);

    props.api
      .getMessages(roomId, { limit: PAGE })
      .then((page) => {
        upsert(page.messages);
        setCursor(page.next_cursor);
      })
      .catch((e) => setError(errorText(e)));

    const sub = props.api.openRoomSocket(roomId, {
      onMessage: (m) => upsert([m]),
      onStatus: (s) => setStatus(s),
    });
    onCleanup(() => sub.close());
  });

  const loadOlder = async () => {
    const c = cursor();
    if (!c) return;
    try {
      const page = await props.api.getMessages(props.roomId, { limit: PAGE, cursor: c });
      upsert(page.messages);
      setCursor(page.next_cursor);
    } catch (e) {
      setError(errorText(e));
    }
  };

  const send = async (e: Event) => {
    e.preventDefault();
    const body = draft().trim();
    if (!body) return;
    setDraft('');
    try {
      const m = await props.api.sendMessage(props.roomId, body);
      upsert([m]);
    } catch (err) {
      setError(errorText(err));
    }
  };

  const addParticipant = async (e: Event) => {
    e.preventDefault();
    const id = Number(participantId());
    if (!Number.isInteger(id) || id <= 0) {
      setError('participant user_id must be a positive number');
      return;
    }
    try {
      await props.api.addParticipant(props.roomId, id);
      setParticipantId('');
      setError(null);
    } catch (err) {
      setError(errorText(err));
    }
  };

  return (
    <section class="room">
      <header>
        <span>Room #{props.roomId}</span>
        <span class="status">{status()}</span>
      </header>
      <Show when={cursor()}>
        <button type="button" class="older" onClick={loadOlder}>
          Load older
        </button>
      </Show>
      <ul class="messages">
        <For each={messages()}>
          {(m) => (
            <li>
              <b class="author">{m.author_handle ?? `#${m.author_id}`}</b>
              <span class="body">{m.body}</span>
              <time>{m.created_at.slice(11, 19)}</time>
            </li>
          )}
        </For>
      </ul>
      <Show when={error()}>
        <p class="err">{error()}</p>
      </Show>
      <form class="send" onSubmit={send}>
        <input
          placeholder="message"
          value={draft()}
          onInput={(e) => setDraft(e.currentTarget.value)}
        />
        <button type="submit">Send</button>
      </form>
      <form class="add-participant" onSubmit={addParticipant}>
        <input
          placeholder="add participant (user_id)"
          value={participantId()}
          onInput={(e) => setParticipantId(e.currentTarget.value)}
        />
        <button type="submit">Add</button>
      </form>
    </section>
  );
}
