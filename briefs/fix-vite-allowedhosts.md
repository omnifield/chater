# chater — Фикс: vite `allowedHosts` (дев-сервер за VS Code-туннелем)

Мелкий фикс фронта. Ветка от свежего main, PR, STOP + отчёт. Ревью — workspace-архитектор + user.

## Проблема (диагностировано)
`vite@5.4.21` имеет host-check (защита от DNS-rebinding): запрос с Host, не входящим в
localhost/allowedHosts, отбивается **403 `Blocked request. This host is not allowed`**. Наш штатный способ
смотреть фронт — **VS Code port-forward**, и при открытии публичной `*.devtunnels.ms`-ссылки Host именно такой
→ vite режет и статику, и прокси `/chater` → в UI «Cannot reach the server». Через `localhost:5173` работает,
но по tunnel-ссылке — нет. Воспроизводится: `curl -H "Host: x.devtunnels.ms" localhost:5173/chater/healthz` → 403.

## Фикс
В `web/vite.config.ts`, блок `server`, добавить:
```ts
// Dev server is viewed through the VS Code port-forward tunnel, whose Host is
// not localhost. This is a LOCAL DEV server (prod path is behind the gateway,
// not this), so allow tunnel hosts. Keep DNS-rebinding note in mind if this ever
// becomes exposed beyond dev.
allowedHosts: true,
```
(Либо точечно: `allowedHosts: ['.devtunnels.ms', 'localhost']` — но VS Code может дать и другой tunnel-домен;
для дев-сервера `true` проще и достаточно. Твой выбор — обоснуй в отчёте.)

## Проверка (DoD)
- `pnpm dev --host`, затем `curl -H "Host: x.devtunnels.ms" localhost:5173/chater/healthz` → **200** (не 403);
  `curl -H "Host: x.devtunnels.ms" localhost:5173/` → 200.
- Открытие фронта по `*.devtunnels.ms`-ссылке из VS Code больше не даёт «Cannot reach the server»
  (комнаты грузятся, Create/Send работают).
- `pnpm -C web build` ок; web-ci зелёный.
- Ветка + PR, STOP → отчёт.

## Опционально (если всплывёт)
Если HMR-live-reload по туннелю ругается на ws — можно добавить `server.hmr.clientPort`/`protocol`.
НЕ обязательно для этого фикса (приложение и API работают и без HMR). Отметь, если увидишь.

## Не в скоупе
Gateway-маршрут / single-origin · прод-конфиг · прочее.
