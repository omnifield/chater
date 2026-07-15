# Бриф: ретайр in-repo CI на reusable + green-proof go-ci

> **Трек:** Foundation — Шаг 1, handoff (ретайр + green-proof go-канона Шага 0)
> **Адресат:** архитектор / owner **chater** (зона: репо chater)
> **Заказчик:** workspace-архитектор (omnifield-hub)
> **Статус:** заказ на исполнение (ветка → PR → CI → ревью → мерж; chater = flow-require-pr)

## North star
Продукт **не несёт инфру — только декларации + product-owned конфиг**. chater перестаёт держать
собственные CI-воркфлоу; тянет reusable из devopser. Меха уже универсальна (Шаг 0/1) — здесь мы её
**потребляем и доказываем**, а не подгоняем.

## Зачем
chater держит **in-repo копии CI** (`.github/workflows/go-ci.yml`, `web-ci.yml`) с NOTE «temporary →
уедет в reusable devopser». Шаг 0 дал reusable `go-ci.yml`, Шаг 1 — стек-осознанный `skeleton init|sync`
+ `ci.yml`-caller per stack. chater объявлен `stack: ["go","frontend"]` в devopser `repo-flow.json`.
**Пора ретайрить копии и доказать go-ci на живом go-модуле** (green-proof, перенесён из Шага 0).

## Скоуп (только репо chater)
1. **Прогнать `skeleton sync`** из devopser: `node <devopser>/packages/skeleton/init.mjs .`
   (или из published-пакета). Материализует общий managed-набор + **`.github/workflows/ci.yml`**
   (caller с двумя job'ами: `go`→go-ci.yml, `node`→node-ci.yml — chater мульти-стек) + `pr-title.yml`.
2. **Ретайрить in-repo воркфлоу:** удалить `.github/workflows/go-ci.yml` и `web-ci.yml` (заменены единым
   `ci.yml`-caller'ом — delete-and-call, не переписывание). Снять NOTE-костыли.
3. **Product-owned оставить как есть:** `.golangci.yml`, `sqlc.yaml` — init-only/product-owned (chater
   правит линтеры/пути/движок БД); reusable go-ci читает ИМЕННО их. Не трогаем содержимое.
4. **Green-proof:** reusable `go-ci` (build·vet·test-race·golangci·sqlc-drift) и `node-ci` проходят
   **зелёными на живом коде chater**. Это проверка механизма Шагов 0–1.

## Green-proof — что именно доказываем (перенос из Шага 0)
Первый реальный go-caller. Если go-ci врёт — вскроется ЗДЕСЬ, до Шага 2. Красные точки-подозреваемые:
`actions/setup-go@v6`, `golangci-lint-action@v7` + `.golangci.yml` schema v2, `tar -xz gitleaks`-member,
sqlc-drift-гейт по наличию `sqlc.yaml`. Красный = баг reusable → **эскалация в devopser** (не костыль в chater).

## DoD (зона chater)
- [ ] `skeleton sync` прогнан; `ci.yml`-caller (go+node) + `pr-title.yml` на месте, идемпотентно.
- [ ] in-repo `go-ci.yml` / `web-ci.yml` удалены; NOTE-костыли сняты; `.golangci.yml`/`sqlc.yaml` сохранены.
- [ ] **reusable go-ci и node-ci зелёные на коде chater** (green-proof); `skeleton drift-check` чист.
- [ ] chater не несёт CI-логику — только `ci.yml`-caller (thin) + product-owned конфиг.
- [ ] PR (require-pr): ветка → CI зелёный → ревью (workspace-архитектор + user) → мерж.

## Handoff (не пункт этого DoD)
- **→ devopser (только если go-ci/node-ci красный на реальном модуле):** эскалация-фидбек с логом; фикс
  reusable — в зоне devopser, не обход в chater.

## Проверка north star (перед мержем)
Если после ретайра chater всё ещё держит собственную CI-логику (не тонкий caller), или green-proof
«зелёный» достигнут костылём/`--no-verify`/выключением гейта — **дефект, не мержим.**
