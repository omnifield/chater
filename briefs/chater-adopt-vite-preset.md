# Бриф: chater на @omnifield/vite-preset + SPA через дверь (Шаг 5, замыкание фронта)

> **Трек:** Foundation — Шаг 5, потребитель vite-пресета
> **Адресат:** архитектор / owner **chater** (зона: репо chater, web/) · chater = flow-require-pr

## North star
chater-фронт открывается на **`:8080/chater`** через дверь. base — из пресета (манифест, единый источник),
ноль хардкода. Пресет `@omnifield/vite-preset@0.1.0` опубликован — тянем под изоляцией.

## Скоуп (репо chater, web/)
1. **Расширить пресет** — `pnpm -C web add -D @omnifield/vite-preset`; `web/vite.config.ts`:
   ```
   import { defineOmnifieldVite } from "@omnifield/vite-preset";
   import solid from "vite-plugin-solid";
   export default defineOmnifieldVite({ plugins: [solid()] /*, server:{proxy:{…}} по нужде*/ });
   ```
   base `/chater/` придёт из манифеста автоматом (пресет). host/allowedHosts — из пресета. Свой vite-хардкод
   base/host убрать (пресет держит).
2. **API через door-контракт:** SPA теперь под `/chater/`, дверь роутит backend на **`/api/chater/`**
   (rewrite→`chater:8020/chater/`). API-клиент фронта (`web/src/api/client.ts`, сейчас зовёт `/chater/rooms`)
   → база **`/api/chater`** (единый путь и в dev, и через дверь). Dev-режим (vite-proxy) привести в
   соответствие: проксировать `/api/chater`→backend (или через пресет-server-оверрайд).
3. **frontend-сервис** (`devbox.services.json`, уже есть) — vite поднимается тем же; при необходимости обновить.

## DoD (зона chater)
- [ ] `web/vite.config.ts` на `defineOmnifieldVite` (base из манифеста, ноль хардкода); `@omnifield/vite-preset` в devDeps.
- [ ] API-клиент через `/api/chater` (door-контракт); dev-proxy согласован.
- [ ] **Живой прог (Канал, после recreate/restart frontend):** `:8080/chater` → **SPA 200**; приложение грузит
      данные (напр. список комнат) через `:8080/api/chater/...` (backend через дверь).
- [ ] `web-ci` (lint/typecheck/test/build) зелёный; PR (require-pr) → ревью → мерж.

## Проверка north star
Хардкод base в vite.config (не из пресета), API мимо door-контракта (`/chater` вместо `/api/chater`), или
`:8080/chater` всё ещё 404 — дефект. Фронт ложится на дверь через пресет, не костылём.
