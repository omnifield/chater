# Бриф: chater декларирует dev-сервис + lifecycle green-proof (Шаг 4)

> **Трек:** Foundation — Шаг 4 (Dev-сервисы lifecycle) — первый реальный потребитель
> **Адресат:** архитектор / owner **chater** (зона: репо chater) · chater = flow-require-pr
> **Заказчик:** workspace-архитектор (omnifield-hub)
> **Статус:** заказ (ветка → PR → CI → ревью → мерж)

## North star
Оркестратор `devbox-services` — **универсальная меха** (existing, skeleton-managed). chater — **первый
реальный потребитель**: декларирует свой сервис и доказывает чистый lifecycle. Продукт-специфика — только
в `devbox.services.json` + путях; меху под chater НЕ подгоняем. Гэп в оркестраторе → **эскалация в devopser**,
не костыль в chater.

## Зачем (факты сняты через Канал)
- `devbox-services.mjs` (skeleton-managed) готов: `up/start/stop/restart/status/run/logs`, pidfile, port-probe
  (G1: сервис обязан слушать **0.0.0.0**, не 127.0.0.1), health. Автостарт завязан (devcontainer
  `postStartCommand: devbox-services up` + Step-2 провижн `run_hook postStartCommand`).
- Схема service-объекта: `{ name, command, port, cwd, healthUrl? }` (оркестратор гонит `sh -c "exec <command>"`).
- chater-backend config-driven: `CHATER_PORT`→Addr (**:8020** prod), `CHATER_DB`→SQLite-путь (**дефолт
  `chater.db` — лёг бы в корень репо = дефект**), goose-миграции идемпотентно на старте, health
  **`/chater/healthz`** уже есть.
- Сейчас `devbox.services.json = []` — backend поднимается ручным `go run`, БД без фикс-пути (каша).

## Скоуп (только репо chater)
1. **`devbox.services.json`** — задекларировать backend-сервис:
   - `name: backend`, `port: 8020`, `cwd: "."`, `healthUrl: "http://localhost:8020/chater/healthz"`.
   - `command`: запуск Go-backend с **инъекцией env** `CHATER_PORT=8020` + `CHATER_DB=<фикс-путь на волюме>`
     (форма — через `env VAR=val …`, т.к. оркестратор оборачивает `exec <command>`; голый `exec VAR=val` не
     сработает). Бинарь слушает **0.0.0.0:8020** (G1), не localhost.
2. **Фикс-путь БД на per-product рантайм-волюме** (MODEL 2.2: рантайм-sqlite → per-product волюм, НЕ репо/git,
   НЕ `/tmp`):
   - добавить в `.devcontainer/devcontainer.json` mount `source=chater-data,target=<напр. /home/vscode/.local/state/chater>,type=volume` (Step-2 провижинер его примонтирует);
   - `CHATER_DB` → файл в этом каталоге. Данные переживают stop→up, docker restart **и `devbox recreate`**.
3. **Ретайр ручного `go run`** — backend поднимается только через `devbox-services`.

## Green-proof lifecycle (честно, на живом backend)
- `devbox-services up` → backend поднят, **health `/chater/healthz` = 200**, bind `0.0.0.0:8020`.
- `devbox-services status` → pid/port/bind/health видны.
- `devbox-services stop` → **порт свободен, зомби-процессов нет** (kill группы).
- Данные: создать запись → `stop`→`up` → запись жива; `docker restart <devbox>` → жива; `devbox recreate` → жива
  (волюм пережил).
- Идемпотентность: повторный `up` = skip, не дубль.

## DoD (зона chater)
- [ ] `devbox.services.json` декларирует backend (port/cwd/health/command с env-инъекцией); bind 0.0.0.0.
- [ ] БД на per-product волюме (манифест + `CHATER_DB`); НЕ в репо, НЕ `/tmp`.
- [ ] lifecycle green-proof пройден: up→healthy, stop→чисто (порт свободен, ноль зомби), данные переживают
      stop→up + restart + recreate, идемпотентно.
- [ ] ручной `go run` ретайрен (backend только через `devbox-services`).
- [ ] PR (require-pr) зелёный → ревью → мерж.

## Handoff / эскалация
- **→ devopser (если оркестратор не тянет реальный Go-сервис):** эскалация с логом (напр. kill-группы `go run`,
  probe-таймаут, health-форма) — фикс в `devbox-services.mjs` (skeleton), не костыль в chater.

## Проверка north star (перед мержем)
Если lifecycle «зелёный» костылём (ручной старт, no-op stop, БД в репо/`/tmp`, bind 127.0.0.1), или меху
подогнали под chater — **дефект, не мержим.** Любой продукт декларацией получает тот же чистый lifecycle.
