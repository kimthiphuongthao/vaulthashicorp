# Project Context (Index)

## TL;DR
Repo này chứa HashiCorp Vault secrets engine plugin `vault-plugin-legacy-cred` (Go) + docker stack để chạy thử và demo.

Plugin hỗ trợ:
- `rotate/<subject>`: sinh plaintext ngẫu nhiên + hash (bcrypt/sha/pbkdf2) + lưu record trong Vault storage
- `updater_type=sql`: update hash ra DB legacy (Postgres/MySQL/MSSQL)
- `updater_type=webhook`: bắn webhook (tuỳ cấu hình)

## Where to start (AI/đồng đội mới)
- Quick context: `docs/ai-context.md`
- Docs index: `docs/README.md`

## Quick run / demo
- Manual demo (SQL standard DBs): `docs/manual-demo-sql-updater-standard-db.md`
- Automated E2E SQL test: `docs/sql-updater-standard-db-test.md`

## Code map
- Plugin entrypoint: `cmd/vault-plugin-legacy-cred/main.go`
- Backend + paths: `internal/legacycred/`
- Updater implementations: `internal/legacycred/updater/`

## Docker deployment model (dev)
- Build + bake plugin binary into Vault image: `docker/Dockerfile.vault-dev`
- Auto register/enable plugin at container start: `docker/entrypoint-vault-dev.sh`

## Companion demo repo (deploy-only bundle)
- Repo demo đã tách riêng (binary + source + compose):
  - `git@github.com:kimthiphuongthao/vault-plugin-legacy-cred-demo.git`

## Common gotchas
- `unsupported path`: thường do gọi nhầm endpoint (SQL config dùng `POST /v1/legacy-cred/config`, không có `/config/sql`).
- `no configuration file provided`: chạy `docker compose` sai thư mục hoặc thiếu `-f`.
