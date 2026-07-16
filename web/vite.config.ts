import { defineOmnifieldVite } from '@omnifield/vite-preset';
import solid from 'vite-plugin-solid';

// Dev backend lives in the same container on the reserved chater port (8020).
// Override with CHATER_BACKEND when needed.
const backend = process.env.CHATER_BACKEND ?? 'http://localhost:8020';

// base (`/chater/`) and the server host/allowedHosts canon come from the preset:
// it derives base from omnifield.yaml's front-route (single source), so nothing
// vite-specific is hardcoded here. We only add the solid plugin, pin the port, and
// stand in for the door with a dev proxy.
export default defineOmnifieldVite({
  plugins: [solid()],
  server: {
    // Pin the port: the reach route in omnifield.yaml (/chater → 5173) must equal
    // the REAL listening port. Vite's default is 5173, but leaving it implicit lets
    // the declared route silently drift; pinning keeps manifest == reality.
    port: 5173,
    proxy: {
      // Door contract: the SPA is served under /chater/, and the door routes the
      // backend at /api/chater (rewrite /api/chater → chater:8020/chater/). In dev
      // the vite proxy stands in for the door — same /api/chater entrypoint,
      // rewritten to the backend's native /chater/ prefix — so the API client uses
      // one path both in dev and behind the door.
      '/api/chater': {
        target: backend,
        changeOrigin: true,
        ws: true,
        rewrite: (path) => path.replace(/^\/api\/chater/, '/chater'),
        // Why: the browser WebSocket API cannot set an Authorization header, but
        // the backend's token-stub auth reads `Authorization: Bearer <handle>`.
        // In dev we let the client pass the handle as a `?token=` query param and
        // move it into the header on the proxied upgrade — the backend stays
        // untouched (header-only). The production browser path (subprotocol /
        // query token at the gateway) arrives with real identity, not now.
        configure: (proxy) => {
          proxy.on('proxyReqWs', (proxyReq, req) => {
            const url = new URL(req.url ?? '', 'http://localhost');
            const token = url.searchParams.get('token'); // URLSearchParams decodes it
            if (token) {
              // Re-encode: header values must be ASCII (a Cyrillic handle would
              // throw here too). The backend percent-decodes it back.
              proxyReq.setHeader('authorization', `Bearer ${encodeURIComponent(token)}`);
            }
          });
        },
      },
    },
  },
});
