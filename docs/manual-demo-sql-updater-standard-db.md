# Demo thủ công SQL updater với DB tiêu chuẩn (Postgres/MySQL/MSSQL)

Mục tiêu: chạy Vault dev + plugin trong container, cấu hình `updater_type=sql`, gọi `rotate/<subject>` và **tự tay kiểm chứng** rằng `password_hash` trong DB đã được cập nhật đúng.

Tài liệu này là phiên bản **manual step-by-step** (không dùng script tự động). Nếu bạn muốn chạy 1 lệnh để test nhanh, xem [docs/sql-updater-standard-db-test.md](docs/sql-updater-standard-db-test.md).

## 0) Prerequisites

- Docker đã chạy
- Repo đã có các service trong [docker-compose.yml](docker-compose.yml)

Quan trọng:
- Bạn phải chạy lệnh trong **thư mục có file** `docker-compose.yml` (root của repo).
- Nếu bạn chạy ở thư mục khác, `docker compose` sẽ báo: `no configuration file provided: not found`.
- Ngoài ra kiểm tra bạn không gõ nhầm `ddocker` (thừa chữ `d`). Lệnh đúng là `docker`.

Nếu không chắc đang ở đúng thư mục, chạy:

```bash
pwd
ls -la
```

Trong môi trường hiện tại, repo root là:

```bash
cd /Volumes/OS/test_failover/vscode
```

Hoặc dùng cách an toàn (chạy ở đâu cũng được) bằng `-f`:

```bash
docker compose -f /Volumes/OS/test_failover/vscode/docker-compose.yml ps
```

Giá trị mặc định trong demo:
- Vault: `VAULT_ADDR=http://127.0.0.1:8200`, token: `root`
- Postgres: user/pass/db = `legacy/legacy/legacy`
- MySQL: user/pass/db = `legacy/legacy/legacy`, root pass = `root`
- MSSQL: `sa/Passw0rd!`, DB `legacy`

## 1) Start stack (Vault + DB)

### 1.1. Start container

Lệnh:

```bash
docker compose up -d --build
```

Nếu bạn không đứng ở repo root, dùng:

```bash
docker compose -f /Volumes/OS/test_failover/vscode/docker-compose.yml up -d --build
```

Đang làm gì:
- Build image `vscode-vault` (Vault dev + plugin baked-in)
- Start các container: Vault, Postgres, MySQL, MSSQL, mock-updater (nếu có)
- Chạy job init cho MSSQL để tạo DB/table

### 1.2. Xem log Vault (để chắc plugin đã mount)

Lệnh:

```bash
docker compose logs --tail=120 vault
```

Nếu bạn không đứng ở repo root, dùng:

```bash
docker compose -f /Volumes/OS/test_failover/vscode/docker-compose.yml logs --tail=120 vault
```

Đang làm gì:
- Xem log startup của Vault
- Bạn nên thấy dòng tương tự: `successful mount ... path=legacy-cred/`

### 1.3. (Optional) Set env cho đỡ gõ dài

Lệnh:

```bash
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
```

Đang làm gì:
- Thiết lập biến môi trường để các lệnh `curl` ngắn hơn

### 1.4. Health check Vault

Lệnh:

```bash
curl -sS "$VAULT_ADDR/v1/sys/health" | cat
```

Đang làm gì:
- Xác nhận Vault đã lên (trả JSON + HTTP 200)

## 2) Kiểm tra endpoint plugin

### 2.1. Đọc config hiện tại

Lệnh:

```bash
curl -sS "$VAULT_ADDR/v1/legacy-cred/config" \
  -H "X-Vault-Token: $VAULT_TOKEN" | cat
```

Đang làm gì:
- Gọi endpoint `legacy-cred/config` để xác nhận mount đã enable và plugin trả dữ liệu

Nếu bạn thấy lỗi `no handler for route ...`, thường là:
- Vault chưa kịp mount plugin (đợi 1-2s và thử lại)
- hoặc container Vault chưa lên hoàn chỉnh (xem log Vault)

## 3) Demo Postgres (cấu hình SQL updater + rotate + verify DB)

### 3.1. Reset hash trong DB (để dễ nhìn thay đổi)

Lệnh:

```bash
docker compose exec -T postgres psql -U legacy -d legacy -v ON_ERROR_STOP=1 \
  -c "UPDATE users SET password_hash='' WHERE username='alice';"
```

Đang làm gì:
- Set `password_hash` về rỗng cho user `alice` trong Postgres

### 3.2. Cấu hình plugin sang SQL updater (Postgres)

Lệnh:

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/config" \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -d '{
    "updater_type":"sql",
    "sql_driver":"postgres",
    "sql_dsn":"postgres://legacy:legacy@postgres:5432/legacy?sslmode=disable",
    "sql_update_query_named":"UPDATE users SET password_hash=:hash WHERE username=:subject",
    "sql_placeholder_style":"dollar"
  }' | cat
```

Đang làm gì:
- Bật updater kiểu `sql`
- Chỉ định driver `postgres` (plugin sẽ map nội bộ sang driver `pgx`)
- DSN trỏ tới service `postgres` trong docker network
- Query dùng placeholder `:hash` và `:subject` (plugin tự convert thành `$1`, `$2`)

### 3.3. Rotate (tạo plaintext + hash + update DB)

Lệnh:

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/rotate/alice" \
  -H "X-Vault-Token: $VAULT_TOKEN" | cat
```

Đang làm gì:
- Plugin generate plaintext ngẫu nhiên
- Hash theo `default_hash_type` (mặc định bcrypt)
- Chạy câu UPDATE vào Postgres để ghi `password_hash`
- Lưu plaintext/hash vào Vault storage (để endpoint `creds` đọc được trong TTL)

### 3.4. Verify trong Postgres

Lệnh:

```bash
docker compose exec -T postgres psql -U legacy -d legacy -tAc \
  "SELECT password_hash FROM users WHERE username='alice';"
```

Đang làm gì:
- Đọc lại `password_hash` từ Postgres
- Đối chiếu với trường `hash` trong response rotate (2 giá trị phải giống nhau)

## 4) Demo MySQL

### 4.1. Reset hash trong MySQL

```bash
docker compose exec -T mysql mysql -ulegacy -plegacy legacy \
  -e "UPDATE users SET password_hash='' WHERE username='alice';"
```

Đang làm gì:
- Reset `password_hash` về rỗng trong MySQL

### 4.2. Cấu hình plugin cho MySQL

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/config" \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -d '{
    "updater_type":"sql",
    "sql_driver":"mysql",
    "sql_dsn":"legacy:legacy@tcp(mysql:3306)/legacy",
    "sql_update_query_named":"UPDATE users SET password_hash=:hash WHERE username=:subject",
    "sql_placeholder_style":"question"
  }' | cat
```

Đang làm gì:
- DSN trỏ tới service `mysql`
- Placeholder style `question` để map thành `?`

### 4.3. Rotate

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/rotate/alice" \
  -H "X-Vault-Token: $VAULT_TOKEN" | cat
```

Đang làm gì:
- Tương tự Postgres: generate → hash → UPDATE MySQL → lưu vào Vault

### 4.4. Verify trong MySQL

```bash
docker compose exec -T mysql mysql -ulegacy -plegacy legacy -Nse \
  "SELECT password_hash FROM users WHERE username='alice';"
```

Đang làm gì:
- Đọc `password_hash` và đối chiếu với `hash` trong response rotate

## 5) Demo MSSQL (SQL Server)

### 5.1. Reset hash trong MSSQL

```bash
docker compose run --rm -T mssql-tools \
  "/opt/mssql-tools/bin/sqlcmd -S mssql -U sa -P \"Passw0rd!\" -d legacy -Q \"SET NOCOUNT ON; UPDATE dbo.users SET password_hash='' WHERE username='alice';\" -b -C"
```

Đang làm gì:
- Reset `password_hash` trong SQL Server

### 5.2. Cấu hình plugin cho MSSQL

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/config" \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -d '{
    "updater_type":"sql",
    "sql_driver":"sqlserver",
    "sql_dsn":"sqlserver://sa:Passw0rd%21@mssql:1433?database=legacy&encrypt=disable",
    "sql_update_query_named":"UPDATE dbo.users SET password_hash=:hash WHERE username=:subject",
    "sql_placeholder_style":"at"
  }' | cat
```

Đang làm gì:
- DSN trỏ tới service `mssql`
- Placeholder style `at` để map thành `@p1`, `@p2` (phù hợp SQL Server)

### 5.3. Rotate

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/rotate/alice" \
  -H "X-Vault-Token: $VAULT_TOKEN" | cat
```

Đang làm gì:
- generate → hash → UPDATE MSSQL → lưu vào Vault

### 5.4. Verify trong MSSQL

```bash
docker compose run --rm -T mssql-tools \
  "/opt/mssql-tools/bin/sqlcmd -S mssql -U sa -P \"Passw0rd!\" -d legacy -Q \"SET NOCOUNT ON; SELECT password_hash FROM dbo.users WHERE username='alice';\" -h -1 -W -b -C"
```

Đang làm gì:
- Đọc `password_hash` và đối chiếu với `hash` trong response rotate

## 6) (Optional) Đọc plaintext từ Vault

Lệnh:

```bash
curl -sS "$VAULT_ADDR/v1/legacy-cred/creds/alice" \
  -H "X-Vault-Token: $VAULT_TOKEN" | cat
```

Đang làm gì:
- Lấy `plaintext` đang được lưu trong Vault storage (dùng cho OpenIG/legacy login)
- Chỉ còn hiệu lực trong TTL

## 7) Cleanup (tắt stack)

```bash
docker compose down -v
```

Đang làm gì:
- Stop container và xóa volume dữ liệu DB (để lần sau demo lại từ đầu, sạch)
