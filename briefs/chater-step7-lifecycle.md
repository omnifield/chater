# chater — Шаг 7: стабилизация dev-lifecycle (адаптация devbox-services)

Цель: чистый **старт/стоп/статус** dev-стека chater, **без зомби-процессов**, **фиксированная БД** (конец
`/tmp/*.db`-каши), переживание рестарта. Делаешь Шаг 7, ветка от свежего main, PR+CI, STOP + отчёт.

## Механизм — АДАПТИРУЕМ существующее (не изобретаем)
В devopser-скелете есть готовый оркестратор `devbox-services.mjs` (zero-deps node): команды
`up / start / stop / restart / status / logs / run`, pidfile, **port-probe**, state в `~/.devbox`
(home = cattle, не пачкает репо). Декларация — `devbox.services.json` в корне репо. Именно это и решает
зомби/ручные пляски. Продукт-фидбек chater-архитектора (нет чистого dev-stop, залипшие :8020) закрывается
ровно этим.

## Скоуп
1. **Вендорни `scripts/devbox-services.mjs`** из devopser-скелета (`packages/skeleton/files/devbox-services.mjs`)
   в chater, с NOTE «синк из devopser-skeleton» (как in-repo CI). Из контейнера скелет недоступен (изоляция) —
   попроси user/workspace-архитектора скопировать файл; я приложу.
2. **`devbox.services.json`** в корне chater:
   - **backend**: команда старта chater (`go run ./cmd/chater` или сборка+бинарь), порт **8020**, bind 0.0.0.0,
     health `/chater/healthz`, **`CHATER_DB` = фиксированный стабильный путь** (напр. `~/.devbox/chater.db` —
     НЕ `/tmp`, НЕ относительный `./`, чтобы данные не разъезжались по файлам).
   - **frontend**: vite (`pnpm -C web dev --host`), порт **5173**, health `/`.
3. **README**: как поднять/остановить/статус dev-стека одной командой (`devbox-services up|stop|status`).

## DoD
- `devbox-services up` → backend(8020)+vite(5173) подняты, `healthz` 200; **`status`** показывает pid/порт/health
  обоих; **`stop`** чисто гасит (порт освобождён, зомби нет); повторный `up` идемпотентен (без дублей).
- Данные **переживают stop→up** (одна фиксированная БД, не новая при каждом старте).
- Ручные `go run` / `pnpm dev` больше не нужны.
- README обновлён. `go`/`web` CI зелёные. Ветка + PR. STOP → отчёт (путь БД, форма services.json, вопросы).

## Не в скоупе (довески после)
- **Автостарт при рестарте контейнера** (контейнер сейчас `sleep infinity`) — нужна правка команды/
  devcontainer devbox'а = инфра/devopser-заход, отдельно.
- Мост brainer как сервис — зона brainer, отдельный бриф.
- Реальная identity, UI-полиш.
