# Architecture — chater

North-star for chater. Kept short; updated as the product grows. The **backend**
(Go) owns the product; a thin **web** UI (`web/`, Vite + Solid) dogfoods it over the
public API — see `web/README.md`.

## What chater is

A generic messenger à la Telegram: **rooms (dialog / group) · messages · history**,
with its **own database**. Language: Go (canon: `~/canon-go.md` in the dev container).

**chater does not know about agents.** The "agent-session = participant" bridge is a
brainer product built on top of chater's *public* API — not our concern. Ecosystem
identity/auth is a separate future product; v0 uses a token-stub.

Public contract: HTTP + websocket under the native prefix `/chater/` (the gateway
proxies `/api/chater/` → `:PORT/chater/`). The listen port is env-only (`CHATER_PORT`)
and the SQLite path is env-only (`CHATER_DB`); no deployment values are hardcoded.

## HTTP API (v0)

JSON over stdlib `net/http` (method+path routing with `{id}`, no framework). All
routes are under `/chater/`.

| Method + path | Purpose | Auth |
|---|---|---|
| `POST /users` | bootstrap a user `{handle}` | none |
| `POST /rooms` | create room `{type, title?, participant_ids?}`; caller becomes owner | Bearer |
| `GET /rooms` | caller's rooms | Bearer |
| `POST /rooms/{id}/participants` | add `{user_id, role?}` → 204 | Bearer, participant |
| `POST /rooms/{id}/messages` | send `{body}`; author = caller | Bearer, **participant (else 403)** |
| `GET /rooms/{id}/messages?limit=&cursor=` | history `{messages, next_cursor}` | Bearer, participant |
| `GET /rooms/{id}/ws` | live room events (websocket) | Bearer, participant |

History pagination returns an **opaque cursor** (base64url of `created_at\x00id`);
callers pass `next_cursor` back to fetch the next (older) page. `next_cursor` is
null on the last page.

### Live delivery (websocket)

`GET /rooms/{id}/ws` upgrades a participant's connection (`coder/websocket`) to a
**receive-only** stream — sending stays on `POST .../messages`, which also
publishes. The membership gate runs *before* the upgrade (non-participant → 403).

Frames are a JSON envelope with room for future event kinds; v0 emits only
`message`, reusing the HTTP message DTO (one wire shape):

```json
{ "type": "message", "message": { …toMessage… } }
```

Fan-out is an in-process `hub` (`room_id → subscribers`, mutex-guarded). `POST`
marshals the event once and publishes **non-blockingly**: a slow/full subscriber
drops frames rather than stalling the broadcast or the HTTP request. Subscribe on
connect, unsubscribe on disconnect/ctx-cancel; the channel is never closed, so
publish stays race-free. Keepalive via periodic server ping.

Known limitation: WS auth reuses the `Bearer` header, which browsers can't set on
a WebSocket — fine for agents/Postman; a browser path (subprotocol/query token)
comes with real identity. Single-instance only; multi-instance pub/sub is later.

### Identity — token-stub (v0)

`Authorization: Bearer <token>` where the token **is** the user's handle; the
middleware resolves it to a user, creating on first use. No signature, secret, or
expiry — a deliberate stub, isolated in `internal/httpapi/auth.go`. Real ecosystem
identity replaces only that file later; handlers receive an already-resolved user.

## Domain (v0)

| Entity | Fields | Notes |
|---|---|---|
| **users** | `id`, `handle` (unique), `created_at` | Minimal identity; no auth providers in v0 |
| **rooms** | `id`, `type` (`dialog`\|`group`), `title?`, `created_at` | Dialog and group are one entity — differ by `type` + participant count |
| **room_participants** | (`room_id`, `user_id`) PK, `role?`, `joined_at` | FKs to rooms/users, `ON DELETE CASCADE` |
| **messages** | `id`, `room_id`, `author_id`, `body`, `created_at` | Index `(room_id, created_at, id)` backs history pagination |

History is paginated **keyset-style** over `(created_at, id)` (newest-first), not by
offset — stable under concurrent inserts.

## Layers

```
cmd/chater/     wiring only: config → open DB → migrate → serve. No logic.
internal/
  config/       env-only configuration (12-factor)
  httpapi/      transport: net/http (stdlib) + coder/websocket, native /chater/
                prefix. Thin handlers, token-stub auth, wire DTOs, in-process
                pub/sub hub for live delivery.
  store/        data access — sqlc-generated types/queries + a thin wrapper.
                Constraint violations surface as ErrNotFound/ErrConflict/
                ErrInvalidReference so handlers map clean 404/409/400.
migrations/     goose SQL migrations (embedded; applied on startup).
```

- **store returns concrete structs**; the *consumer* (httpapi) declares the narrow
  interface it needs (canon: accept interfaces, return structs). This keeps the DB
  swappable.
- **Types are generated from the SQL schema** via sqlc — no hand-written row→struct
  mapping. Queries live in `internal/store/queries.sql`; sqlc reads the schema from the
  goose migrations (single source of truth).
- **Migrations run on startup** (idempotent `goose up`). Rationale: dev-first, one
  moving part, always schema-ready. A standalone `migrate` command can be split out
  later if prod ops want migrations gated separately from rollout.
- **DB**: SQLite for dev via the pure-Go `modernc.org/sqlite` driver (no cgo; clean
  under `-race`). SQL is kept reasonably portable; Postgres is a later drop-in — a
  driver + DDL swap behind the `store` boundary, not a call-site rewrite.

## Non-goals (v0)

Real auth/identity providers · attachments · reactions · search · threads ·
presence/typing · WS-send (v0 websocket is receive-only) · multi-instance pub/sub.
These are the Telegram-parity / scale horizon, explicitly out of v0.

## Build / CI

In-repo go-ci gate: `gofmt` · `go vet` · `go test -race` · `golangci-lint` · sqlc-drift
check. Tool versions are pinned (see `.github/workflows/go-ci.yml`). **NOTE:** this gate
and tool pins migrate to the reusable devopser go-ci later (precedent: weber).
