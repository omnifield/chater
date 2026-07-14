import { ApiError } from './api/client';
import type { Room } from './api/types';

export function roomLabel(r: Room): string {
  return r.title ?? `${r.type} #${r.id}`;
}

/** Turn any thrown value into a human-readable line for the UI. */
export function errorText(e: unknown): string {
  if (e instanceof ApiError) {
    switch (e.status) {
      case 0:
        return 'Cannot reach the server.';
      case 401:
        return 'Not authenticated — set a handle.';
      case 403:
        return 'You are not a participant of this room.';
      case 404:
        return 'Not found.';
      default:
        return e.message;
    }
  }
  if (e instanceof Error) return e.message;
  return 'Unknown error';
}
