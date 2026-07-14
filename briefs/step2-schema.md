# chater — Шаг 2: схема + миграции (goose + sqlc) + ARCHITECTURE.md

Шаг 1 замержен (PR #3). Продолжаешь ты, chater-архитектор, сам. Делаешь **ТОЛЬКО Шаг 2**,
ветка + PR, потом STOP + отчёт. Ревью — workspace-архитектор + user.

## Скоуп Шага 2 — только слой данных
БД-слой: миграции создают схему, sqlc генерит типы/запросы, накат с нуля чистый. **HTTP-ручек
комнат/сообщений ещё НЕ делаешь** (это Шаг 3) — только store + его тесты. Go-канон под рукой: `~/canon-go.md`.

## Сущности v0 (по founding-брифу)
- **users** — `id`, `handle` (уникальный), `created_at`. Минимальная identity; auth-провайдеров нет
  (token-stub позже, отдельно).
- **rooms** — `id`, `type` (`dialog` | `group`), `title` (nullable), `created_at`. Диалог и группа —
  ОДНА сущность, различие в `type` + числе участников.
- **room_participants** — `room_id`, `user_id`, `role` (nullable), `joined_at`. PK — составной
  (`room_id`,`user_id`). FK на rooms/users.
- **messages** — `id`, `room_id` (FK), `author_id` (FK users), `body`, `created_at`.
  **Индекс `(room_id, created_at)`** — под пагинацию истории.

## Технически (канон)
- **SQLite** (dev), **goose**-миграции в `migrations/` (up/down), **sqlc** (types-from-schema).
  Инструменты go-устанавливаемые (`go install` goose/sqlc) — зафиксируй версии.
- Доступ к БД — в `internal/` (напр. `internal/store`), **интерфейс у потребителя** (канон), реализация
  за ним → потом Postgres drop-in без переписывания вызовов.
- sqlc-запросы минимум: create user; create room; add participant; insert message; **list messages by
  room с пагинацией** (keyset по `(created_at,id)` предпочтительно); list rooms пользователя.
- Накат: реши и задокументируй — миграции при старте сервиса ИЛИ отдельная команда `migrate` (env-only конфиг).

## ARCHITECTURE.md (заводим здесь)
Короткий north-star продукта: что такое chater (generic-мессенджер, своя БД, **про агентов не знает**);
доменные сущности (выше); слоистость (`cmd`/`internal`/`migrations`, store за интерфейсом);
не-цели v0 (auth-провайдеры, вложения/реакции/поиск/треды — горизонт). Лаконично, без воды.

## DoD Шага 2
- goose-миграции создают схему; `goose up` с пустой БД — чисто; `down` работает.
- sqlc генерит; запросы компилируются; `go build ./... && go vet ./... && go test -race ./...` зелёные.
- **table-driven тесты стора** против SQLite (`-race`): вставка/выборка, пагинация истории.
- `ARCHITECTURE.md` заведён.
- Ветка + PR, CI зелёный. STOP → отчёт (решения по миграции-на-старте vs команда, версии goose/sqlc, вопросы).

## НЕ в Шаге 2
HTTP-ручки комнат/сообщений (Шаг 3) · websocket (Шаг 4) · auth · фронт · фабрикация devcontainer/
devbox-session (ждём Go-скелет от devopser).
