# chater web

Minimal chat UI (Vite + Solid + TypeScript) over the live chater API. Not a mock —
it talks to the real backend.

## Layout

```
web/
  index.html
  vite.config.ts       # dev server + /chater proxy (HTTP + ws token bridge)
  vitest.config.ts     # jsdom + solid plugin
  biome.json           # lint/format
  src/
    main.tsx           # mount
    App.tsx            # login gate + rooms/room layout
    api/
      client.ts        # ApiClient — the single typed seam (HTTP + ws). ChatApi interface.
      types.ts         # wire types (mirror backend DTOs)
      index.ts         # shared, env-configured client instance
    state/session.ts   # token-stub handle in localStorage
    components/        # Login, Rooms, RoomView
    format.ts          # room labels + human error text
```

Components depend on the `ChatApi` interface, never on `fetch`/`WebSocket` directly —
`ApiClient` implements it; tests pass a stub.

## Run (dev)

Backend and frontend both live in this container.

```sh
# 1) backend on the reserved chater port, with a dev DB
(cd .. && CHATER_PORT=8020 CHATER_DB=./chater-dev.db go run ./cmd/chater)

# 2) frontend
pnpm install
pnpm dev            # http://localhost:5173
```

Open the vite URL via the VS Code port forward. Enter a handle to "log in"
(token-stub — the handle is your `Bearer` token; the first request creates the user).

- API base is configurable via `VITE_API_BASE` (default: same origin → vite proxy).
- Backend target for the proxy is `CHATER_BACKEND` (default `http://localhost:8020`).

### Live over websocket

The browser `WebSocket` API can't set an `Authorization` header, so the client sends
the handle as `?token=<handle>` and the **vite dev proxy** moves it into the header on
the upgrade request — the backend stays header-only. In production this is replaced by
real identity (subprotocol / gateway token); the backend is untouched either way.

## Checks

```sh
pnpm lint         # biome
pnpm typecheck    # tsc --noEmit
pnpm test         # vitest
pnpm build        # vite build
```

> NOTE: the tool pins and `web-ci.yml` gate are in-repo for now and migrate to the
> reusable devopser skeleton later (same as `go-ci`).
