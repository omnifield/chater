// Token-stub session: the "token" is just the user's handle, kept in
// localStorage. No real auth (v0) — this is the single place the handle lives.

const KEY = 'chater.handle';

export function loadHandle(): string | null {
  return localStorage.getItem(KEY);
}

export function saveHandle(handle: string): void {
  localStorage.setItem(KEY, handle);
}

export function clearHandle(): void {
  localStorage.removeItem(KEY);
}
