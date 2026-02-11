# AI Context — Vault legacy-cred plugin (repo gốc)

Tài liệu này giúp AI agent/đồng đội hiểu nhanh repo: mục tiêu, cách chạy, API, config schema thực tế, và nơi tìm scripts/docs.

## 1) Mục tiêu
Plugin Vault secrets engine `vault-plugin-legacy-cred`:
- Sinh plaintext ngẫu nhiên (password) theo cấu hình.
- Hash plaintext (bcrypt/sha/pbkdf2) theo cấu hình.
- Lưu record trong Vault storage theo `subject` + `version` + `expires_at`.
- Đồng bộ hash ra hệ thống legacy qua `updater_type`:
  - `noop`
  - `webhook`
  - `sql` (Postgres/MySQL/MSSQL) — cập nhật `password_hash` cho user.

## 2) Cách chạy nhanh (dev)
- Stack dev có sẵn trong docker compose: Vault + DBs.
- Điểm quan trọng: plugin được **bake vào image Vault dev**.

Xem hướng dẫn:
- Manual demo SQL updater: [docs/manual-demo-sql-updater-standard-db.md](docs/manual-demo-sql-updater-standard-db.md)
- Automated SQL updater test: [docs/sql-updater-standard-db-test.md](docs/sql-updater-standard-db-test.md)
- Curl test tổng quan: [docs/curl-test.md](docs/curl-test.md)

## 3) Deployment model trong repo gốc (build + register + enable)
- Build plugin binary: [docker/Dockerfile.vault-dev](docker/Dockerfile.vault-dev)
  - stage `builder` build Go binary `vault-plugin-legacy-cred`
  - stage vault copy vào `/vault/plugins/vault-plugin-legacy-cred`
- Khi container start, entrypoint sẽ auto:
  - start Vault dev với `-dev-plugin-dir=/vault/plugins`
  - tính SHA256 plugin
  - `vault plugin register ...`
  - `vault secrets enable -path=legacy-cred vault-plugin-legacy-cred`
  - set default config

Xem script: [docker/entrypoint-vault-dev.sh](docker/entrypoint-vault-dev.sh)

## 4) API surface (đúng endpoint)
- Config:
  - `GET /v1/legacy-cred/config`
  - `POST /v1/legacy-cred/config`
- Rotate:
  - `POST /v1/legacy-cred/rotate/<subject>`
- Read record:
  - `GET /v1/legacy-cred/creds/<subject>`

Lưu ý: Cấu hình SQL updater **không** dùng `/config/sql`. SQL config được set qua `POST /v1/legacy-cred/config` với các field `sql_*`.

## 5) Scripts/tests quan trọng
- Automated SQL updater E2E: [scripts/test-sql-updater.sh](scripts/test-sql-updater.sh)
  - chờ Vault ready
  - cấu hình plugin -> rotate -> query DB -> so sánh hash

## 6) Troubleshooting
- `no configuration file provided: not found`: chạy `docker compose` sai thư mục hoặc thiếu `-f`.
- `unsupported path`: thường do gọi nhầm endpoint (vd `/config/sql`).
- Apple Silicon warning (amd64 image): thường không ảnh hưởng demo.

## 7) Repo demo (bundle tối giản)
Repo demo tách riêng để deploy nhanh: `kimthiphuongthao/vault-plugin-legacy-cred-demo`.
