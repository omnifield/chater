# chater — Шаг 3: HTTP API (rooms / messages) + token-stub identity

Шаг 2 замержен (`main` ae39f5a). Продолжаешь ты, chater-архитектор, сам. **ТОЛЬКО Шаг 3**,
ветка + PR, потом STOP + отчёт. Ревью — workspace-архитектор + user (user гоняет **Postman'ом**).
Go-канон: `~/canon-go.md`.

## Скоуп Шага 3 — HTTP-слой поверх store
Ручки под нативным префиксом `/chater/`, JSON, поверх `internal/store`. **Websocket ещё НЕТ** (Шаг 4) —
живые сообщения потом; сейчас history через HTTP.

## Identity (token-stub, честно как заглушка)
- Auth-провайдеров нет. Простой стаб: `Authorization: Bearer <token>` → резолв в user
  (в v0 токен = handle или id пользователя; middleware достаёт/создаёт user). Документируй как СТАБ,
  вынеси в одно место — заменится на реальную identity позже, не переписывая ручки.
- `POST /chater/users {handle}` → создать пользователя (bootstrap для тестов).

## Эндпоинты (форма; допускается уточнить, но покрыть эти сценарии)
- `POST /chater/rooms` `{type, title?, participant_ids?}` → room. Создатель — участник автоматически.
- `GET  /chater/rooms` → комнаты текущего пользователя (ListRoomsForUser).
- `POST /chater/rooms/{id}/participants` `{user_id, role?}` → 204. Добавить участника.
- `POST /chater/rooms/{id}/messages` `{body}` → message. Автор = текущий пользователь;
  **требование: он участник комнаты, иначе 403.**
- `GET  /chater/rooms/{id}/messages?limit=&cursor=` → `{messages, next_cursor}`.
  Курсор — непрозрачная строка (кодируй `(created_at,id)`); отдавай `next_cursor` для след. страницы.

## Канон / требования
- stdlib `net/http`, method+path паттерны с `{id}` (Go 1.22 mux). Без фреймворка.
- **Интерфейс у потребителя:** `httpapi` объявляет узкий интерфейс стора, который ему нужен;
  `*store.Store` его удовлетворяет. Не тащи весь store в хендлеры.
- Валидация входа; лимит на размер тела; корректные коды (400/403/404/409/500); errors `%w`; slog.
- Авторизация реальная: отправка сообщения/чтение истории — только участник комнаты.
- Провязка в `cmd/chater`: open DB (`CHATER_DB` env, путь; для dev допусти файл) → `store.Migrate` на старте →
  `store.NewStore` → router со стором. Конфиг env-only (добавь `CHATER_DB`).

## DoD Шага 3
- Эндпоинты выше работают; `go build ./... && go vet ./... && go test -race ./...` зелёные.
- **table-driven httptest-тесты** (`-race`): happy-path каждого эндпоинта + auth-стаб + **enforcement участника
  (403)** + пагинация истории через HTTP.
- Сервис поднимается (CHATER_PORT + CHATER_DB), миграции на старте; **проверяемо Postman'ом** end-to-end:
  создать 2 юзеров → комнату → добавить участника → послать сообщения → получить историю страницами.
- Ветка + PR, CI зелёный (вкл. sqlc-drift). STOP → отчёт (форма эндпоинтов, решения по token-stub/курсору, вопросы).

## НЕ в Шаге 3
Websocket (Шаг 4) · реальная identity/auth · фронт · фабрикация devcontainer/devbox-session.
