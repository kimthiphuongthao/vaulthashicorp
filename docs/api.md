# API & Endpoints

Giả sử enable secrets engine tại mount path `legacy-cred/`.

## 1) `legacy-cred/config`

### Read
- Operation: `read`
- Mục đích: xem config hiện tại (DSN sẽ được mask)

Ví dụ:

```bash
vault read legacy-cred/config
```

### Update
- Operation: `update`
- Mục đích: set/override cấu hình

Các field chính:

- `default_ttl` (seconds): TTL plaintext (mặc định 10800)
- `password_length` (int): độ dài plaintext generate
- `password_charset` (string): `alnum|ascii|<custom-alphabet>`
- `one_time_read` (bool): nếu true, đọc `creds` xong sẽ xoá record

Hash:
- `default_hash_type`: `bcrypt|sha256|sha512|pbkdf2`
- `bcrypt_cost`
- `sha_salt`, `sha_salt_position` (`prefix|suffix`), `sha_output_encoding` (`hex|base64`)
- `pbkdf2_iterations`, `pbkdf2_key_length`, `pbkdf2_salt_length`, `pbkdf2_hash` (`sha256|sha512`)

Updater:
- `updater_type`: `noop|webhook|sql`
- Webhook: `webhook_url`, `webhook_method`, `webhook_headers` (JSON string)
- SQL: `sql_driver`, `sql_dsn`, `sql_update_query_named`, `sql_placeholder_style`

Ví dụ (noop):

```bash
vault write legacy-cred/config \
  default_ttl=10800 \
  default_hash_type=bcrypt \
  bcrypt_cost=12 \
  updater_type=noop
```

## 2) `legacy-cred/rotate/<subject>`

- Operation: `update`
- Mục đích: rotate password cho 1 subject

Input (optional):
- `hash_type`: override hash type riêng cho lần rotate
- `ttl` (seconds): override TTL plaintext cho subject

Ví dụ:

```bash
vault write legacy-cred/rotate/user@example.com
```

Response hiện tại trả về cả `plaintext` và `hash` để test nhanh. (Trong production có thể chuyển sang response-wrapping/giảm bề mặt lộ plaintext.)

## 3) `legacy-cred/creds/<subject>`

- Operation: `read`
- Mục đích: OpenIG đọc plaintext hiện hành

Ví dụ:

```bash
vault read legacy-cred/creds/user@example.com
```

Hành vi:
- Nếu `expires_at` đã qua: plugin xoá record và trả empty
- Nếu `one_time_read=true`: đọc xong xoá record
