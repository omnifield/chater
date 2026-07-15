import {
  createEffect,
  createResource,
  createSignal,
  For,
  onCleanup,
  onMount,
  Show,
} from 'solid-js';
import type { ChatApi } from '../api/client';
import type { Room, RoomType } from '../api/types';
import { errorText, roomLabel } from '../format';

const DEFAULT_POLL_MS = 6000;

/**
 * mergeRooms produces the next visible list from the freshly-fetched one, reusing
 * the previous object for an unchanged room so <For> (keyed by reference) keeps
 * that row's DOM node — new/changed rooms appear, gone rooms drop, and there is
 * no flicker of the whole list.
 */
export function mergeRooms(prev: Room[], fresh: Room[]): Room[] {
  const byId = new Map(prev.map((r) => [r.id, r] as const));
  return fresh.map((fr) => {
    const existing = byId.get(fr.id);
    if (existing && existing.title === fr.title && existing.type === fr.type) {
      return existing;
    }
    return fr;
  });
}

/** Rooms lists the caller's rooms (live-polled) and lets them create a new one. */
export function Rooms(props: {
  api: ChatApi;
  selectedId: number | null;
  onSelect: (id: number) => void;
  pollMs?: number;
}) {
  const [initial, { refetch }] = createResource(() => props.api.listRooms());
  const [rooms, setRooms] = createSignal<Room[]>([]);
  const [loadError, setLoadError] = createSignal<string | null>(null);
  const [type, setType] = createSignal<RoomType>('group');
  const [title, setTitle] = createSignal('');
  const [createError, setCreateError] = createSignal<string | null>(null);

  // Drive the visible list from the resource via non-throwing accessors
  // (.latest / .error) — reading the value accessor would throw on error/pending
  // and could tear down the app.
  createEffect(() => {
    if (initial.error) {
      setLoadError(errorText(initial.error));
      return;
    }
    const data = initial.latest;
    if (data) {
      setLoadError(null);
      setRooms((prev) => mergeRooms(prev, data));
    }
  });

  // Poll so a new room / invite (incl. being added by someone else) shows up
  // without a page reload. Soft-merge keeps rows stable; selection lives in the
  // parent, so it is never disturbed. A transient poll failure keeps the list.
  onMount(() => {
    const timer = setInterval(async () => {
      try {
        const fresh = await props.api.listRooms();
        setRooms((prev) => mergeRooms(prev, fresh));
        setLoadError(null);
      } catch {
        // keep current list; retry next tick
      }
    }, props.pollMs ?? DEFAULT_POLL_MS);
    onCleanup(() => clearInterval(timer));
  });

  const create = async (e: Event) => {
    e.preventDefault();
    setCreateError(null);
    try {
      const room = await props.api.createRoom(type(), title().trim() || undefined);
      setTitle('');
      await refetch();
      props.onSelect(room.id);
    } catch (err) {
      setCreateError(errorText(err));
    }
  };

  const retry = () => {
    setLoadError(null);
    void refetch();
  };

  return (
    <aside class="rooms">
      <h2>Rooms</h2>

      {/* Create form is always mounted, above the list — interactive regardless
          of the list's load state. */}
      <form class="create-room" onSubmit={create}>
        <select value={type()} onChange={(e) => setType(e.currentTarget.value as RoomType)}>
          <option value="group">group</option>
          <option value="dialog">dialog</option>
        </select>
        <input
          placeholder="title (optional)"
          value={title()}
          onInput={(e) => setTitle(e.currentTarget.value)}
        />
        <button type="submit">Create</button>
      </form>
      <Show when={createError()}>
        <p class="err">{createError()}</p>
      </Show>

      <Show
        when={loadError()}
        fallback={
          <ul>
            <For each={rooms()}>
              {(r) => (
                <li>
                  <button
                    type="button"
                    classList={{ active: r.id === props.selectedId }}
                    onClick={() => props.onSelect(r.id)}
                  >
                    {roomLabel(r)}
                  </button>
                </li>
              )}
            </For>
          </ul>
        }
      >
        <div class="err">
          <p>Could not load rooms: {loadError()}</p>
          <button type="button" onClick={retry}>
            Retry
          </button>
        </div>
      </Show>
    </aside>
  );
}
