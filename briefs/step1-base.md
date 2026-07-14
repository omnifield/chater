# chater backend — Шаг 1: база (Go-проект по канону + бегущий health-сервис)

**Ты** — chater-архитектор. Поднимаешь chater сам, без owner-сессий пока. Ревью —
workspace-архитектор + user **ПОСЛЕ КАЖДОГО шага**. Делаешь **ТОЛЬКО Шаг 1**, потом STOP и отчёт.
Вперёд не убегай.

## Продукт
Generic-мессенджер (комнаты / сообщения / история), **своя БД**, Go. **Про агентов НЕ знает**
(agent-as-participant — мост поверх публичного API, позже). Вижн — прочитай первыми:
`briefs/founding-backend-v0.md`, `README.md`, `ARCHITECTURE.md`.

## Правила
- ⚠️ **main закрыт require-PR ruleset'ом** (прямой push отбивается GH013). Работаешь в **ветке**,
  открываешь **PR** — это точка ревью.
- **Честная база:** скелет devopser node-ориентированный; готового **Go-скелета НЕТ**. Возьми
  репо-агностичное из скелета (editorconfig, gitattributes, gitignore + Go-блок, devcontainer,
  `devbox-session.sh`), а **Go-проект разложи по Go-канону сам, в самом chater**, с in-repo go-ci и
  **NOTE «уедет в devopser»** (founding-бриф разрешает, прецедент weber). Позже канон-часть перевозим
  в devopser — держи её обособленной, чтобы вынос был `git mv`, не переписывание.
- **Go-канон:** `knowledger/standards/canon/languages/go.md`. Из контейнера он недоступен (изоляция) —
  **попроси user скопировать** его в chater, если нужен полный текст.

## Go-канон — ключевое для Шага 1
- `go.mod` с **toolchain-пином**; раскладка `cmd/chater/` (main) + `internal/` (логика) + `migrations/`.
- **stdlib-first HTTP** (`net/http`, без фреймворка); errors `%w` + context-first; `slog`; **env-only** конфиг.
- table-driven тесты + `go test -race`; `golangci-lint` в CI-гейте.

## DoD Шага 1
- `go.mod` + канон-раскладка; `go build ./...` зелёный.
- **Health:** `GET /chater/healthz → 200`, под нативным префиксом `/chater/`.
- Сервис стартует в `chater-devbox`; **порт — из env** (канон env-only). Реальная аллокация из
  `devopser registry/ports.md` и gateway-маршрут — **позже, вне Шага 1**.
- in-repo **go-ci** (build / vet / test -race / golangci-lint) + NOTE «уедет в devopser».
- **Ветка + PR открыт.** STOP → отчёт: что сделал, ключевые решения, вопросы.

## НЕ делаешь на Шаге 1
БД/миграции · комнаты/сообщения · websocket · фронт (его вообще пока нет — тест через Postman позже).
**Только база.** Шаг 2 (схема+миграции) — по нашему go после ревью.
