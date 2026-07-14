import { createSignal } from 'solid-js';
import { saveHandle } from '../state/session';

/** Login collects a handle and stores it as the token-stub identity. */
export function Login(props: { onLogin: (handle: string) => void }) {
  const [handle, setHandle] = createSignal('');

  const submit = (e: Event) => {
    e.preventDefault();
    const h = handle().trim();
    if (!h) return;
    saveHandle(h);
    props.onLogin(h);
  };

  return (
    <form class="login" onSubmit={submit}>
      <h1>chater</h1>
      <p>Enter a handle to start — it becomes your identity (token-stub, no password in v0).</p>
      <input
        placeholder="handle"
        value={handle()}
        onInput={(e) => setHandle(e.currentTarget.value)}
      />
      <button type="submit">Enter</button>
    </form>
  );
}
