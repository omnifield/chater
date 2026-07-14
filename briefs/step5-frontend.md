# chater — Шаг 5: фронт (Vite + Solid) — минимальный чат-UI поверх живого API

Backend v0 готов и в `main` (538843f). Продолжаешь ты, chater-архитектор. **ТОЛЬКО Шаг 5**,
ветка от свежего main (сразу!), PR, потом STOP + отчёт. Ревью — workspace-архитектор + user (смотрит в браузере).
Go-канон не нужен; фронт-канон — Solid (shared-policy §3: DOM/Solid, jsdom).

## Что делаем
Первый **видимый** чат: залогиниться, видеть комнаты, открыть комнату, читать историю, писать,
**получать сообщения live**. Против РЕАЛЬНОГО API (бэкенд уже есть — НЕ мок).

## Стек (решено, не выбираешь)
**Vite + Solid + TypeScript**, тесты **Vitest + jsdom**, линт/формат **Biome** (канон/скелет).
Живёт в `web/` внутри репо chater (свой `package.json`, pnpm). Минимальные стили, без UI-фреймворка.

## Экраны (минимум)
1. **Identity (token-stub):** ввод `handle` → это твой `Bearer <handle>`; сохрани в localStorage.
   Реальной auth нет. Первый запрос сам создаёт юзера (бэкенд `GetOrCreateUserByHandle`).
2. **Rooms:** список (`GET /chater/rooms`) + создание (`POST /chater/rooms {type,title}`; создатель — участник).
3. **Room view:**
   - история: `GET /chater/rooms/{id}/messages?limit=&cursor=` + «загрузить старее» (пагинация по `next_cursor`);
   - отправка: `POST /chater/rooms/{id}/messages {body}`;
   - **live:** `GET /chater/rooms/{id}/ws` → на кадр `{type:"message",message:{…}}` добавляй сообщение в ленту.
   - (мин.) добавить участника — поле `user_id` (по handle→id API-эндпоинта нет; не расширяй бэкенд сейчас,
     достаточно user_id или `participant_ids` при создании комнаты).

## Требования
- **Типизированный API-клиент — единый шов** (`ApiClient` модуль): один слой над HTTP+ws, компоненты не
  дёргают fetch напрямую. База API — **конфигурируема** (env/vite), НЕ хардкод.
- **Dev-провязка:** `vite.config` proxy `/chater` → `http://localhost:8020` (бэкенд в этом же контейнере),
  `ws: true` для ws-эндпоинта. Запуск: бэкенд (`CHATER_PORT=8020 CHATER_DB=./chater-dev.db`) + `vite dev`.
  User смотрит через **VS Code port-forward** порта vite (портов наружу не публикуем).
- Обработка ошибок API (401/403/4xx) в UI по-человечески; ws-reconnect при обрыве (простой).
- Тесты **Vitest**: API-клиент (против замоканного fetch) + ключевые компоненты (рендер комнат, отправка).
  **Biome** чисто; `vite build` проходит.

## DoD Шага 5
- `web/` Solid-app: login по handle → список/создание комнат → комната с историей (пагинация) + отправка +
  **live-обновления по ws**. Против реального бэкенда.
- Типизированный `ApiClient` (единый шов, конфигурируемая база); компоненты через него.
- Запускается: бэкенд + `vite dev`; **видно в браузере через VS Code forward**; live работает
  (открыть комнату в двух вкладках → сообщение появляется в обеих).
- Vitest зелёный, Biome чисто, build ок. CI-джоб для фронта добавлен (pnpm i / biome / vitest / build)
  с NOTE «уедет в devopser-скелет» (как go-ci).
- Ветка + PR, CI зелёный. STOP → отчёт (структура web/, решения по клиенту/proxy/reconnect, вопросы).

## НЕ в Шаге 5
Реальная auth · presence/typing · вложения/реакции/поиск · причёсанный дизайн · gateway-маршрут
(single-origin подвяжем отдельно) · фабрикация devcontainer.
