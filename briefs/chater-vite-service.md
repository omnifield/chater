# Бриф: chater декларирует vite как dev-сервис (хвост Шага 5 — фронт /chater)

> **Трек:** Foundation — Шаг 5, хвост (фронт `/chater` через дверь)
> **Адресат:** архитектор / owner **chater** (зона: репо chater) · chater = flow-require-pr

## North star
`:8080/chater` отдаёт живой фронт (сейчас 502 — маршрут двери верен, но vite не бежит). vite — второй
dev-сервис chater под тем же оркестратором (Шаг 4), не ручной запуск.

## Скоуп (репо chater)
1. **Добавить frontend-сервис в `devbox.services.json`** (рядом с backend):
   - `name: frontend`, `cwd: web`, `command: "pnpm dev"` (vite; `web/vite.config.ts` `host:true` → bind
     **0.0.0.0:5173** для G1; `CHATER_BACKEND` дефолтит на `http://localhost:8020`), `port: 5173`,
     `healthUrl: "http://localhost:5173/"`.
2. `devbox recreate` (хост) — на старте `devbox-services up` поднимет оба (backend+frontend); заодно свежий
   `pnpm -C web install` активирует esbuild-фикс (`onlyBuiltDependencies`, PR #20 — vite получит native-бинарь).

## DoD (зона chater)
- [ ] `devbox.services.json`: frontend-сервис (vite, port 5173, bind 0.0.0.0, health).
- [ ] `devbox-services status` → backend+frontend оба up/healthy; ноль ручных запусков/зомби.
- [ ] **Живой прог (Канал, после recreate):** `:8080/chater` → **200** (фронт через дверь), backend `/api/chater` не сломан.
- [ ] PR (require-pr) зелёный → ревью → мерж.

## Проверка north star
Ручной vite вместо сервиса, bind 127.0.0.1 (G1 провал), или /chater всё ещё 502 после recreate — дефект.
