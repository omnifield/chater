import { createResource, createSignal, For, Show } from 'solid-js';
import type { ChatApi } from '../api/client';
import type { RoomType } from '../api/types';
import { errorText, roomLabel } from '../format';

/** Rooms lists the caller's rooms and lets them create a new one. */
export function Rooms(props: {
  api: ChatApi;
  selectedId: number | null;
  onSelect: (id: number) => void;
}) {
  const [rooms, { refetch }] = createResource(() => props.api.listRooms());
  const [type, setType] = createSignal<RoomType>('group');
  const [title, setTitle] = createSignal('');
  const [error, setError] = createSignal<string | null>(null);

  const create = async (e: Event) => {
    e.preventDefault();
    setError(null);
    try {
      const room = await props.api.createRoom(type(), title().trim() || undefined);
      setTitle('');
      await refetch();
      props.onSelect(room.id);
    } catch (err) {
      setError(errorText(err));
    }
  };

  return (
    <aside class="rooms">
      <h2>Rooms</h2>
      {/* When the resource errors, reading rooms() rethrows — so guard the list
          behind the error check and never read the value on the error branch. */}
      <Show
        when={!rooms.error}
        fallback={<p class="err">Could not load rooms: {errorText(rooms.error)}</p>}
      >
        <ul>
          <For each={rooms() ?? []}>
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
      </Show>
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
      <Show when={error()}>
        <p class="err">{error()}</p>
      </Show>
    </aside>
  );
}
