import { loadHandle } from '../state/session';
import { ApiClient } from './client';

// The single shared client. Base URL is configurable (env), defaulting to
// same-origin so the vite dev proxy fronts both HTTP and websocket.
export const api = new ApiClient({
  baseUrl: import.meta.env.VITE_API_BASE ?? '',
  getToken: () => loadHandle(),
});

export type { ChatApi } from './client';
export { ApiError } from './client';
