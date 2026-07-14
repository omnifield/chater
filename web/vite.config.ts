import { defineConfig } from 'vite';
import solid from 'vite-plugin-solid';

// Dev backend lives in the same container on the reserved chater port (8020).
// Override with CHATER_BACKEND when needed.
const backend = process.env.CHATER_BACKEND ?? 'http://localhost:8020';

export default defineConfig({
  plugins: [solid()],
  server: {
    proxy: {
      '/chater': {
        target: backend,
        changeOrigin: true,
        ws: true,
        // Why: the browser WebSocket API cannot set an Authorization header, but
        // the backend's token-stub auth reads `Authorization: Bearer <handle>`.
        // In dev we let the client pass the handle as a `?token=` query param and
        // move it into the header on the proxied upgrade — the backend stays
        // untouched (header-only). The production browser path (subprotocol /
        // query token at the gateway) arrives with real identity, not now.
        configure: (proxy) => {
          proxy.on('proxyReqWs', (proxyReq, req) => {
            const url = new URL(req.url ?? '', 'http://localhost');
            const token = url.searchParams.get('token');
            if (token) {
              proxyReq.setHeader('authorization', `Bearer ${token}`);
            }
          });
        },
      },
    },
  },
});
