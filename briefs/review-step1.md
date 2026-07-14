# Ревью — Шаг 1 (chater base). Вердикт: принято по существу, НЕ мержим — 1 фикс

**От:** workspace-архитектор + user. **PR:** #3 (`step1-base`).
**Статус:** код принят, но **CI красный (`lint`)** → мерж после фикса ниже.

## Что проверено (прогнано, не на слово)
- ✅ `go build` / `go vet` / `go test -race` — зелёные.
- ✅ **Health живьём:** `GET /chater/healthz` → `200 {"status":"ok"}`; неизвестный путь → `404`; чистый slog-старт.
- ✅ Код канон-формы: graceful shutdown (SIGINT/SIGTERM), `ReadHeaderTimeout`, `errors.Is(ErrServerClosed)`,
  `%w`, context-first, slog JSON, env-only конфиг, native-префикс `/chater/`, method-prefixed роуты,
  `run()` отделён и тестируем. Раскладка канон (`cmd/`+`internal/`+`migrations/`, toolchain-пин). Претензий нет.
- ❌ CI: `build-test` PASS, **`lint` FAIL**.

## БЛОКЕР до мержа (1 строка)
Причина lint-фейла — **не версия линтера и не схема**, а **версия экшна**. Из лога дословно:
> `golangci-lint v2 is not supported by golangci-lint-action v6, you must update to golangci-lint-action v7`

**Фикс:** в `.github/workflows/go-ci.yml` подними `golangci/golangci-lint-action@v6` → **`@v7`**.
v7 поддерживает golangci-lint v2 + твою v2-схему `.golangci.yml` и пин `v2.1.6` — их НЕ трогай.
Перезапусти CI; если после этого всплывут реальные находки линтера — поправь (код чистый, риск низкий).
Коммить в тот же PR #3.

## Ответы на 4 вопроса
1. **devcontainer/`devbox-session.sh` — не фабрикуй, правильно, что не стал.** Слепая копия без исходника =
   дрейф = костыль. Для бэка не нужны (запуск через `docker exec`). Это продукт-фидбек devopser'у
   (Go-скелет-гэп) — логируем; приедут с Go-скелетом / при переносе канон-части. Ручные
   `.editorconfig/.gitattributes/.gitignore` — ок.
2. **Go-канон дан полным:** лежит в контейнере — **`~/canon-go.md`** (`/home/vscode/canon-go.md`, вне репо,
   git не засоряет). Строгий сверочный проход по Шагу 1 не нужен (проверено вручную). Держи под рукой для
   Шага 2 — там важны конвенции goose/sqlc/migrations.
3. **ARCHITECTURE.md — ошибка в step1-брифе (спутал с brainer).** У chater его нет, для Шага 1 не нужен.
   Заведём на **Шаге 2** вместе со схемой/доменом.
4. **golangci — см. блокер:** проблема в версии экшна, не линтера. Пин `v2.1.6` оставь, экшн → `@v7`.

## Дальше
1. Фикс `action v6→v7` в PR #3 → CI зелёный.
2. Сообщи workspace-архитектору «CI green» → подтверждаю **мерж PR #3**.
3. Получаешь **Шаг 2** (схема + миграции: goose + sqlc; сущности `users` / `rooms` / `room_participants` /
   `messages`; + завести `ARCHITECTURE.md`).
