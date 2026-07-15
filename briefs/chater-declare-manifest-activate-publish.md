# Бриф: chater декларирует omnifield.yaml + активирует publish (Шаг 5.2)

> **Трек:** Foundation — Шаг 5.2 (продукт декларирует манифест-визитку + вживляет publish-volume)
> **Адресат:** архитектор / owner **chater** (зона: репо chater) · chater = flow-require-pr
> **Заказчик:** workspace-архитектор (omnifield-hub)
> **Статус:** заказ (ветка → PR → CI → ревью → мерж)

## North star
chater — первый продукт, открывающийся через **одну дверь :8080** из своего манифеста. Манифест —
**декларативная визитка** (product-owned контракт `@omnifield/contract-manifest`), не скрейп. Publish в
`omnifield-registry` активируется на этом продукте (Шаг 5.1 дал меху). Ноль port-forward-обходов.

## Зачем (факты сняты через Канал)
- Шаг 5.1 дал `scripts/devbox-publish.mjs` (managed, вендорнут в chater) + template mount/postStart — но
  **devcontainer.json существующего chater init-only, ещё старый** → volume `omnifield-registry` не смонтирован,
  publish не запускается.
- `omnifield.yaml` у chater **нет** (только weber имеет). Без него chater — вне двери.
- Роутинг chater (registry/ports.md): backend Go **:8020**, нативный префикс `/chater/`; фронт vite
  (`web/`, `host:true`, внутри проксит `/chater`-API на :8020). Gateway-маршруты: `/chater/` (фронт) +
  `/api/chater/` (backend). generate.mjs мапит `reach.routes[{path,port}]` → nginx `location path → chater:port`.

## Скоуп (только репо chater)
1. **`omnifield.yaml`** в корне (форма — как `weber/omnifield.yaml`, контракт `omnifield.dev/v1`):
   `name: chater`, `type`, `title`, `description`, `integration.deps` (если есть), и **`reach.routes`**:
   - фронт: `{ path: /chater, port: <vite-порт из web/vite.config> }`;
   - backend-api: `{ path: /api/chater, port: 8020 }`.
   - **Conform `registry/ports.md`** (порты внутренние; :8020 — контрактный). `reach.port` = РЕАЛЬНЫЙ listening-порт.
2. **Активировать publish** — синкнуть свой `.devcontainer/devcontainer.json` к template Шага 5.1
   (init-only, вручную): добавить mount `source=omnifield-registry,target=/omnifield-registry,type=volume`,
   `chown` в postCreate, postStart `node scripts/devbox-publish.mjs; node scripts/devbox-services.mjs up`.
3. **`devbox recreate`** (хостовый — mount-изменение требует пересоздания, Шаг 2) → на старте chater публикует
   `omnifield.yaml` → `omnifield-registry/chater.yaml`.

## Coupling с 5.3 (флаг, НЕ скоуп этого брифа)
Маршрут `/api/chater/` → backend, который слушает под нативным `/chater/` → генератор (hub-core, 5.3) обязан
**переписать префикс** `/api/chater/` → `chater:8020/chater/`. Текущий generate.mjs делает голый `proxy_pass`
без rewrite. **chater декларирует маршрут честно (`/api/chater` → 8020); rewrite — зона hub-core (5.3).**
Если 5.3 ещё не готов — chater всё равно публикует манифест (проверяется публикацией в volume), а живой
`:8080/api/chater` подтвердится после 5.3.

## DoD (зона chater)
- [ ] `omnifield.yaml` валиден по контракту, conform `registry/ports.md` (routes: /chater→vite, /api/chater→8020).
- [ ] `devcontainer.json` синкнут (mount `omnifield-registry` + chown + postStart publish); `devbox recreate` прошёл.
- [ ] **Живой прог (Канал):** после старта chater-devbox в `omnifield-registry/chater.yaml` лежит манифест
      (байт-идентичен корневому); `devbox-publish` в логе `✓`.
- [ ] PR (require-pr) зелёный → ревью → мерж.

## Handoff
- **→ hub-core (5.3):** появился второй валидный манифест (chater) + маршрут с префикс-rewrite (`/api/chater`).
  Генератор должен его собрать + rewrite. Сквозной `:8080/chater` — DoD 5.3.

## Проверка north star (перед мержем)
Манифест-скрейп вместо декларации, порт≠реальный listening, publish не активирован (volume не смонтирован),
или продукт торчит мимо :8080 — **дефект, не мержим.**
