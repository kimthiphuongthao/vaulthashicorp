# Hashing & Updater

## Hashing

Mặc định: `bcrypt` (phổ biến trong nhiều hệ thống legacy).

### bcrypt
- Output là chuỗi dạng `$2a$...`/`$2b$...`
- Config: `bcrypt_cost`

### sha256 / sha512
- Output encoding: `hex` (default) hoặc `base64`
- Optional salt:
  - `sha_salt`: string
  - `sha_salt_position`: `prefix|suffix`

### pbkdf2
- Mỗi lần rotate sẽ random salt mới.
- Output là chuỗi tự mô tả:

`pbkdf2$<sha256|sha512>$<iters>$<saltB64>$<dkB64>`

Phù hợp để thử nghiệm (nhưng nếu legacy app của bạn yêu cầu format khác, ta sẽ phải chỉnh format output cho đúng).

## Updater (cập nhật legacy credential store)

Plugin hỗ trợ 3 chế độ:

### 1) `noop`
- Không update gì.
- Dùng để test logic rotate/creds trong Vault.

### 2) `webhook` (khuyến nghị khi thử nghiệm đa DB)
- Plugin gọi HTTP endpoint do bạn tự triển khai.
- Ưu điểm: DB platform nào cũng test được (Oracle/MySQL/Postgres/…); plugin không cần nhúng driver.

Config ví dụ:

```bash
vault write legacy-cred/config \
  updater_type=webhook \
  webhook_url="https://legacy-updater.example.com/update" \
  webhook_method=POST \
  webhook_headers='{"Authorization":"Bearer ..."}'
```

Payload:

```json
{"subject":"user@example.com","hash":"...","hash_type":"bcrypt","version":1,"updated_at":"2026-02-10T00:00:00Z"}
```

### 3) `sql` (kết nối DB trực tiếp)

Config cần:
- `sql_driver`: `postgres` | `mysql` | `sqlserver`
- `sql_dsn`: connection string
- `sql_update_query_named`: query dùng placeholder `:hash` và `:subject`
- `sql_placeholder_style`: `question` hoặc `dollar` hoặc `at`

Gợi ý:
- Postgres: `dollar` (ra `$1`, `$2`)
- MySQL: `question` (ra `?`)
- MSSQL: `at` (ra `@p1`, `@p2`)

Ví dụ Postgres:

```bash
vault write legacy-cred/config \
  updater_type=sql \
  sql_driver=postgres \
  sql_dsn="postgres://user:pass@host:5432/db?sslmode=disable" \
  sql_update_query_named='UPDATE users SET password_hash=:hash WHERE username=:subject' \
  sql_placeholder_style=dollar
```

Lưu ý:
- SQL updater hiện mở/đóng connection theo mỗi rotate (đơn giản cho POC). Nếu cần tối ưu, có thể thêm pool/reuse.
