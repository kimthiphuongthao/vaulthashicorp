# SQL updater với DB tiêu chuẩn (Postgres/MySQL/MSSQL)

Tài liệu này mô tả chính xác những gì vừa được thực hiện: bổ sung DB chuẩn trong docker-compose, thêm init schema, cập nhật plugin để hỗ trợ MSSQL placeholder, và thêm script test tự động để xác nhận hash được ghi vào DB.

## 1) Các thay đổi chính

### 1.1. SQL placeholder style `at` cho MSSQL
- Bổ sung placeholder style `at` để map `:hash/:subject` thành `@p1/@p2` (phù hợp SQL Server).
- Cho phép `sql_placeholder_style=at` trong config.

File cập nhật:
- [internal/legacycred/updater/updater.go](internal/legacycred/updater/updater.go)
- [internal/legacycred/paths_config.go](internal/legacycred/paths_config.go)

### 1.2. DB chuẩn trong docker-compose
Thêm các service DB chuẩn để test trực tiếp SQL updater:
- `postgres:16-alpine`
- `mysql:8.4`
- `mssql:2022-latest` + init job

File cập nhật:
- [docker-compose.yml](docker-compose.yml)

### 1.3. Init schema chuẩn cho cả 3 DB
Schema thống nhất: bảng `users(username, password_hash)`.

Files mới:
- [docker/db-init/postgres/01_init.sql](docker/db-init/postgres/01_init.sql)
- [docker/db-init/mysql/01_init.sql](docker/db-init/mysql/01_init.sql)
- [docker/db-init/mssql/init.sql](docker/db-init/mssql/init.sql)

### 1.4. Script test tự động end-to-end
Script sẽ:
1) Start Vault + DBs
2) Chờ DB sẵn sàng
3) Cấu hình `legacy-cred/config` sang `updater_type=sql`
4) `rotate` và verify `password_hash` trong DB khớp với `hash` trả về

File mới:
- [scripts/test-sql-updater.sh](scripts/test-sql-updater.sh)

### 1.5. Cập nhật docs liên quan
- Bổ sung hướng dẫn test SQL updater vào tài liệu curl test.
- Nêu rõ placeholder style cho MSSQL.

Files cập nhật:
- [docs/curl-test.md](docs/curl-test.md)
- [docs/hashing-and-updater.md](docs/hashing-and-updater.md)

## 2) Thông số dùng trong test

### Vault
- `VAULT_ADDR`: `http://127.0.0.1:8200`
- `VAULT_TOKEN`: `root`
- mount path: `legacy-cred/`

### DB credentials
- Postgres: `legacy/legacy`, DB `legacy`
- MySQL: `legacy/legacy`, DB `legacy` (root password `root`)
- MSSQL: `sa/Passw0rd!`, DB `legacy`

### DSN & placeholder style
- Postgres
  - `sql_driver=postgres`
  - `sql_dsn=postgres://legacy:legacy@postgres:5432/legacy?sslmode=disable`
  - `sql_placeholder_style=dollar`
- MySQL
  - `sql_driver=mysql`
  - `sql_dsn=legacy:legacy@tcp(mysql:3306)/legacy`
  - `sql_placeholder_style=question`
- MSSQL
  - `sql_driver=sqlserver`
  - `sql_dsn=sqlserver://sa:Passw0rd%21@mssql:1433?database=legacy&encrypt=disable`
  - `sql_placeholder_style=at`

## 3) Kết quả test (đã xác nhận)
- Postgres: `password_hash` được update đúng bằng `hash` trả về.
- MySQL: `password_hash` được update đúng bằng `hash` trả về.
- MSSQL: `password_hash` được update đúng bằng `hash` trả về.

## 4) Cách chạy lại nhanh

```bash
chmod +x scripts/test-sql-updater.sh
./scripts/test-sql-updater.sh
```

Kết quả mong đợi: log in ra `All SQL updater tests passed.`
