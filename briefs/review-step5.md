# Ревью — Шаг 5 (chater web / Solid). Вердикт: ПРИНЯТ, замержен (PR #7)

**От:** workspace-архитектор + user.

## Проверено (код + живьём + build)
- ✅ CI зелёный (go build-test + web build-test + lint).
- ✅ **Живой data-path UI:** поднял бэкенд :8020 + `vite dev` :5173; через **vite-proxy** прогнал
  healthz → создать юзера → комнату → сообщение → история. Ровно путь, которым ходит UI.
- ✅ `vite build` → dist, 19.7 kB JS (8 kB gzip), 282ms.
- ✅ Код: `ChatApi`-интерфейс = единый шов (компоненты не трогают fetch), configurable base,
  ws-reconnect с backoff, **дедуп по id в `upsert`** (эхо своего POST не дублирует live-кадр),
  корректный Solid-lifecycle (resubscribe + onCleanup). vite.config красиво решает браузерный ws-auth
  (`?token=` → заголовок на прокси, бэкенд не тронут).

Блокеров нет.

## Ответы на 3 вопроса
1. **vite5/vitest2 standalone + override `vite:5.4.21` вместо эталонных vite8/vitest4 — ок.** Эталон (weber) —
   nx-моно с оверрайдами; для чистого standalone стабильная пара 5/2 разумнее. Версии подравняем, когда
   у devopser появится фронт-пресет/скелет.
2. **repo-local `.pnpm-store` (169M) в .gitignore — верно.** Корень: volume-стор
   (`~/.local/share/pnpm/store`) на другой ФС, чем bind-репо → pnpm делает локальный стор рядом с проектом.
   **Да, это продукт-фидбек devopser'у** (devbox/skeleton: store-dir для вложенных JS-пакетов на той же ФС,
   что репо / общий стор). Зафиксировано. Пока gitignore — правильный immediate-fix.
3. **biome `recommended:true` (deprecation-info, не фейл), `biome migrate` НЕ гонять (уводит в preset:none =
   выключает линт) — ок.** До появления biome-пресета экосистемы держим так. Адаптируем пресет, когда выйдет.

## Итог
**chater v0 = бэкенд + лицо.** Есть работающий чат: вход по handle → комнаты → история/отправка → live.
Прод-форма, свой стек, тесты, CI. Готов к догфуду.
