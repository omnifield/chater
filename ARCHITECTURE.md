# Architecture — chater

North-star for the chater backend. Kept short; updated as the product grows.

## What chater is

A generic messenger à la Telegram: **rooms (dialog / group) · messages · history**,
with its **own database**. Language: Go (canon: `~/canon-go.md` in the dev container).

**chater does not know about agents.** The "agent-session = participant" bridge is a
brainer product built on top of chater's *public* API — not our concern. Ecosystem
identity/auth is a separate future product; v0 uses a token-stub.

Public contract: HTTP + websocket under the native prefix `/chater/` (the gateway
proxies `/api/chater/` → `:PORT/chater/`). The listen port is env-only (`CHATER_PORT`);
no deployment port is hardcoded.

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
  httpapi/      transport: net/http (stdlib), native /chater/ prefix. Thin handlers.
  store/        data access — sqlc-generated types/queries + a thin wrapper.
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

Auth providers · attachments · reactions · search · threads. These are the Telegram-parity
horizon, explicitly out of v0. HTTP room/message handlers land in Step 3, websocket in Step 4.

## Build / CI

In-repo go-ci gate: `gofmt` · `go vet` · `go test -race` · `golangci-lint` · sqlc-drift
check. Tool versions are pinned (see `.github/workflows/go-ci.yml`). **NOTE:** this gate
and tool pins migrate to the reusable devopser go-ci later (precedent: weber).
