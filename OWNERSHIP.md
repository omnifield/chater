# OWNERSHIP — chater

Обстановка для **chater-архитектора** (main-сессия этого репо). Прочитать ПЕРВЫМ,
до founding-брифа. Обновлять при смене обстановки — это living-док.

## Что такое chater

Generic-мессенджер а-ля telegram: комнаты (диалог/группа) / сообщения / история.
Своя БД. Язык — **Go** (решение user 2026-07-08). Публичный контракт: HTTP +
websocket, нативный префикс `/chater/`.

**Chater про агентов НЕ знает.** Мост «агент-сессия = участник чата» — продукт
brainer, поверх нашего публичного API, не наша зона. Identity/auth экосистемы —
отдельный будущий продукт: в v0 токен-стаб.

## Двойная миссия

1. **Продукт** — по `briefs/founding-backend-v0.md` (скоуп v0, DoD).
2. **Полигон чистого агент-воркфлоу**: chater строится агентами с нуля; каждый шаг —
   живой тест мехов экосистемы (брифы, PR-флоу, brainer-пульт). Находка про сам
   воркфлоу = продукт-фидбек соответствующему владельцу (brainer/devopser/оракул)
   брифом через user — чужое у себя не чинить.

## Обстановка (2026-07-11)

| Что | Состояние |
|---|---|
| Репо | README-стаб + `briefs/` + этот файл. Кода нет — намеренно. Ничего не подтягивать руками из других реп |
| Git-флоу | **main ЗАКРЫТ** — ruleset `flow-require-pr` (проверен живьём): только ветки+PR; squash-merge, заголовок PR = conventional commit |
| CI | Пока НЕТ. Первый шаг кода = in-repo CI с NOTE «уедет в reusable go-ci devopser» (прецедент weber; план devopser — `devopser/briefs/chater-go-prereqs.md` §П.2) |
| Скелет | Go-вариант у devopser в плане (hooksPath+sh хуки без node: pre-commit `gofmt -l`+`go vet`, pre-push `go test -race`); до его выката — минимум руками ПО ЭТОМУ ЖЕ плану, чтобы синк потом был тривиален |
| Порт | **8020** застолблён (`devopser/registry/ports.md`); gateway-маршрут `/api/chater/` появится с runtime — заказ devopser'у тогда же |
| Среда | Containers-only: рабочая копия — клон в WSL2 FS, всё исполняется в devbox (`ghcr.io/omnifield/devbox`, go внутри, toolchain докачается по go.mod); креды — общий volume `omnifield-secrets` (уже заполнен, заносить нечего) |

## Канон (источники, не пересказы)

- **Go**: `knowledger/standards/canon/languages/go.md` — раскладка cmd/internal/migrations,
  go.mod toolchain-пин, stdlib-first HTTP + coder/websocket, goose+sqlc, slog/env-only,
  table-driven + `-race`.
- **Containers-only**: `devopser/briefs/containers-only-and-management.md` + `devopser/devbox/README.md`
  (пути входа, занос кредов, известное поведение).
- **Процесс**: этапная разработка, проверка после этапа → коммит; брифы двухуровневые
  (TL;DR user'у + спека исполнителю); WIP=1.

## Флоу работы

Роадмап-ход (user+ты) → бриф → owner-агент (user запускает через brainer-пульт) →
PR → ревью (ты) → merge. Эскалация: продукт-вопросы → user; кросс-репо/канон →
оракул через user.

## Границы зон

| Зона | Владелец |
|---|---|
| Код/архитектура chater | ты (+owner-агенты по брифам) |
| Ruleset/CI-reusable/скелет/порты | devopser |
| Пульт/спавн сессий/мост participant↔session | brainer |
| Канон-статьи | knowledger (правки через оракула) |
| Кросс-продуктовые контракты | оракул |
