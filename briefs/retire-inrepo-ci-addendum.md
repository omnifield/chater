# Addendum: полный ретайр chater (go + web) — разблокировано

> **Дополняет:** `briefs/retire-inrepo-ci-onto-reusable.md` (go-only, был красный по green-proof)
> **Трек:** Foundation — Шаг 1, закрытие handoff chater
> **Адресат:** архитектор / owner **chater** · **Зона:** репо chater · chater = flow-require-pr

## Что изменилось (разблокировка)
green-proof поймал 2 бага reusable go-ci + дырку frontend-стека. Оба закрыты в devopser:
- **go-ci пофикшен** (PR #7): sqlc-пин v1.31.1, golangci `install-mode: goinstall`.
- **web-ci построен + frontend-стек разведён** (PR #8): standalone-фронт → `web-ci` (не node-ci),
  корневой nx-набор фронту не навязывается. `repo-flow.json` уже даёт chater
  `stack:["go","frontend"]` + `frontend.working-directory: "web"`.

Теперь ретайр go **и** web закрывается одним PR.

## Важно: прошлый sync был баговый — сбросить начисто
Текущая ветка держит **незакоммиченный налив старого init** (frontend→node): ошибочные корневые
`nx.json`/`package.json`/`biome.json`/`.github/dependabot.yml` и `ci.yml`, сгенерённый со СТАРЫМ поведением
(go+node job вместо go+web). Не коммитить это.

## Скоуп (репо chater, один PR)
1. **Сброс багового налива:** откатить/почистить незакоммиченные артефакты прошлого sync
   (`git checkout -- .editorconfig .gitattributes .gitignore`; удалить ошибочные untracked
   `nx.json` `package.json` `biome.json` `.github/dependabot.yml` — chater НЕ node-монорепо).
   Легитимный общий набор (`.husky/` `.npmrc` `.devcontainer/` `devbox.services.json` `scripts/`) — оставить.
2. **Перепрогнать `skeleton sync`** с текущего devopser `@main` (frontend разведён):
   `node <devopser>/packages/skeleton/init.mjs .` → `ci.yml` теперь с job'ами **`go`→go-ci + `web`→web-ci
   (`with: working-directory: web`)** + `pr-title.yml`. Проверить: **корневых `nx.json`/`package.json`/
   `biome.json` НЕТ**.
3. **Пины фронт-воркспейса:** добавить в `web/package.json` `packageManager` (pnpm@<канон>) + `engines.node`
   (<канон>) — web-ci берёт версии ИМЕННО отсюда (сейчас их нет → пробилдит хардкод-факты, канон требует пины).
4. **Ретайр in-repo воркфлоу:** удалить `.github/workflows/go-ci.yml` и `web-ci.yml` (заменены единым
   `ci.yml`-caller'ом — delete-and-call). Снять NOTE-костыли. `.golangci.yml`/`sqlc.yaml` — product-owned, оставить.
5. **Green-proof (обе ветви честно-зелёные на живом коде):**
   - **go-ci** (build·vet·test-race·golangci[goinstall]·sqlc-drift[v1.31.1]) — зелёный.
   - **web-ci** (`web/`: install·lint·typecheck·test·build) — зелёный.

## DoD (зона chater)
- [ ] баговый налив сброшен; корневых `nx/package/biome` нет; общий набор + `ci.yml`(go+web)+`pr-title` на месте.
- [ ] in-repo `go-ci.yml`/`web-ci.yml` удалены; `.golangci.yml`/`sqlc.yaml` сохранены; NOTE сняты.
- [ ] `web/package.json` несёт `packageManager` + `engines.node` (пины воркспейса).
- [ ] **go-ci И web-ci зелёные** на коде chater (green-proof); `skeleton drift-check` чист.
- [ ] chater не несёт CI-логику — только `ci.yml`-caller + product-owned конфиг.
- [ ] PR (require-pr) → CI зелёный → ревью (workspace-архитектор + user) → мерж.

## Handoff
- **→ devopser (только если reusable красный на реальном коде):** эскалация с логом; фикс — в devopser.

## Проверка north star (перед мержем)
Зелень green-proof костылём/`--no-verify`/no-op'ом = дефект. chater обязан остаться **тонким**:
caller + product-owned, ноль собственной CI-логики, ноль корневого nx-хлама.
