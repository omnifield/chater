# Ответы к Шагу 5 (фронт) — до старта

## 1. Фронт-канон + эталонные конфиги — приложены
Лежит в контейнере: **`~/ref-frontend.txt`** (`/home/vscode/ref-frontend.txt`). Внутри (разделители `=====FILE:...=====`):
- **`shared-policy.md`** (в т.ч. §3 — DOM/Solid, jsdom) — канон.
- **weber `biome.json`** — эталон линт/формат-конфига экосистемы (возьми за основу своего `web/biome.json`).
- **weber `tsconfig.base.json`** — TS-конвенции.
- **weber `packages/ui/vitest.config.ts`** + **`ui/package.json`** — как настроен Solid + jsdom + solid-плагин
  и какие версии deps (`solid-js`, `@solidjs/testing-library`, `vite-plugin-solid`, `vitest`, `jsdom`).

⚠️ **Это эталон КОНВЕНЦИЙ, не СТРУКТУРЫ.** weber — тяжёлый nx-монорепо; **не копируй** его apps/core/packages/
nx.json. chater `web/` = **один маленький Vite+Solid-app**, без nx, без монорепо. Бери из эталона: biome-правила,
tsconfig-строгость, vitest+jsdom+solid-плагин setup, версии deps. `vite.config.ts` пиши сам (стандартный
`vite-plugin-solid` + `server.proxy`). Полноценный фронт-скелет в devopser отсутствует (как и Go-скелет) —
твой `web/` c NOTE «уедет в devopser-скелет» и есть его зачаток.

## 2. Порт / single-origin — да, подтверждаю
Дев-бэкенд на **8020** (застолблён в registry) ✓. Фронт ходит через `vite proxy` на `localhost:8020` в том же
контейнере. **Gateway/single-origin в Шаге 5 НЕ трогаем** — маршрут `/api/chater/` подвяжем отдельным заходом.
Ты прав.

## 3. CI — отдельный `web-ci.yml` с path-фильтром
Заводи **отдельный workflow `.github/workflows/web-ci.yml`**, триггер `on: {push,pull_request: {paths: ['web/**']}}`
(pnpm install / biome check / vitest / vite build). Отдельный — потому что тулчейн и триггеры независимы от Go;
проще уедет в devopser-скелет. NOTE «уедет в devopser» — так же, как go-ci. (go-ci можно позже тоже сузить
по `paths`, но это не в этом шаге.)

Старт разрешён. Ветка от свежего main сразу.
