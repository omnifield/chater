# Founding-brief — chater backend v0 (Go)

| | |
|---|---|
| **Адресат** | chater-архитектор (сессии агентов user запускает через brainer-пульт) |
| **От** | оракул-архитектор, 2026-07-11 (решения user) |
| **Контекст** | chater = полигон чистого агент-воркфлоу: строится агентами С НУЛЯ; каждый шаг разработки = живой тест мехов (брифы, PR-флоу, пульт brainer) |
| **Пререквизиты** | заказы devopser'у отданы (`devopser/briefs/chater-go-prereqs.md`): require-PR ruleset + Go-вариант скелета/CI. До go-ci допустим in-repo CI с NOTE «уедет в devopser» (прецедент weber) |

## Что такое chater (границы продукта)

Generic-мессенджер а-ля telegram: **комнаты (диалог и группа — одна сущность
с участниками) / сообщения / история**. СВОЯ БД. **Про агентов НЕ знает** —
мост «агент-сессия = participant» делает brainer у себя, позже, поверх ПУБЛИЧНОГО
API chater. Телеграм-паритет (вложения/реакции/поиск/тред) = горизонт, НЕ v0.

## Скоуп v0 (минимум, ничего сверх)

1. **Сущности**: user (минимальный: id+handle) · room (участники, тип диалог/группа) ·
   message (текст, автор, ts). БЕЗ auth-провайдеров (локальный токен-стаб; identity
   экосистемы — отдельный продукт, потом).
2. **API**: HTTP (rooms CRUD-минимум: create/list/join; messages: send/history
   с пагинацией) + **websocket** на живые сообщения комнаты (coder/websocket,
   канон Go).
3. **БД**: SQLite для dev, goose-миграции + sqlc (types-from-schema) — Postgres
   drop-in потом.
4. **Контракт стабилен снаружи**: URL-префикс `/chater/` нативный (гейт-parity
   как у brainer: gateway проксирует `/api/chater/` → `:PORT/chater/`; порт —
   заказать в `devopser registry/ports.md`, хардкодов в коде нет).

## Канон (обязателен, не пересказан — читать источники)

- **Go-канон**: `knowledger/standards/canon/languages/go.md` — go.mod toolchain-пин,
  раскладка `cmd/` `internal/` `migrations/`, stdlib-first HTTP, errors `%w` +
  context-first, интерфейс у потребителя, slog + env-only конфиг, table-driven
  тесты + `-race`, CI-гейт.
- **Containers-only**: весь движняк в контейнере (devbox умеет go); рабочая копия —
  клон в WSL2 FS.
- **Git-флоу**: main ЗАКРЫТ, ветки + PR с первого дня (ruleset devopser),
  conventional-commits, squash-merge = заголовок PR.

## Флоу работы (это и есть тест)

Роадмап-ход → бриф → owner-агент через brainer-пульт → PR → ревью → merge.
Находки про сам ВОРКФЛОУ (неудобство пульта, дыры брифов, гэпы скелета) — это
продукт-фидбек: brainer'у / devopser'у / оракулу брифами через user, не чинить
молча у себя.

## DoD v0

Сервис поднимается в контейнере; две комнаты + группа, сообщения текут по ws,
история с пагинацией через HTTP; goose-миграции применяются с нуля; тесты
table-driven + `-race` зелёные; CI зелёный; префикс-parity (нативный `/chater/`)
готов к gateway-маршруту.
