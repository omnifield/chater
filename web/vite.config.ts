import { defineConfig } from 'vite';
import solid from 'vite-plugin-solid';

// Dev backend lives in the same container on the reserved chater port (8020).
// Override with CHATER_BACKEND when needed.
const backend = process.env.CHATER_BACKEND ?? 'http://localhost:8020';

export default defineConfig({
  plugins: [solid()],
  server: {
    // Pin the port: the gateway route in omnifield.yaml (/chater → 5173) must equal
    // the REAL listening port. Vite's default is 5173, but leaving it implicit lets
    // the declared route silently drift; pinning keeps manifest == reality.
    port: 5173,
    // Bind all interfaces (0.0.0.0), not just localhost. Vite defaults to
    // IPv6 localhost (::1); a devcontainer/WSL port-forward reaches the server
    // over IPv4 (127.0.0.1), so an ::1-only bind makes the browser hang on an
    // endless load. host:true makes it reachable over both.
    host: true,
    // The dev server is viewed through the VS Code port-forward, whose Host is a
    // rotating `*.devtunnels.ms` subdomain — vite's DNS-rebinding host check
    // (403 "This host is not allowed") otherwise blocks both static assets and
    // the /chater proxy. This is a LOCAL DEV server only (production traffic is
    // fronted by the gateway, never this), and the tunnel domain isn't fixed, so
    // allow any host rather than pinning one. Revisit if this is ever exposed
    // beyond dev.
    allowedHosts: true,
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
