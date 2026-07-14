import { createSignal, ErrorBoundary, Show } from 'solid-js';
import { api } from './api';
import { Login } from './components/Login';
import { Rooms } from './components/Rooms';
import { RoomView } from './components/RoomView';
import { clearHandle, loadHandle } from './state/session';

export function App() {
  const [handle, setHandle] = createSignal<string | null>(loadHandle());
  const [roomId, setRoomId] = createSignal<number | null>(null);

  const logout = () => {
    clearHandle();
    setRoomId(null);
    setHandle(null);
  };

  return (
    <Show when={handle()} fallback={<Login onLogin={(h) => setHandle(h)} />}>
      {/* Safety net: any uncaught throw in the authed tree becomes a recoverable
          screen instead of a frozen, dead app. */}
      <ErrorBoundary
        fallback={(err, reset) => (
          <div class="app-error">
            <p>Something went wrong: {err instanceof Error ? err.message : String(err)}</p>
            <button type="button" onClick={reset}>
              Reload
            </button>
          </div>
        )}
      >
        <div class="app">
          <div class="topbar">
            <span>
              chater — <b>{handle()}</b>
            </span>
            <button type="button" onClick={logout}>
              Logout
            </button>
          </div>
          <div class="layout">
            <Rooms api={api} selectedId={roomId()} onSelect={setRoomId} />
            <Show
              when={roomId()}
              fallback={<section class="room empty">Pick or create a room</section>}
            >
              {(id) => <RoomView api={api} roomId={id()} />}
            </Show>
          </div>
        </div>
      </ErrorBoundary>
    </Show>
  );
}
