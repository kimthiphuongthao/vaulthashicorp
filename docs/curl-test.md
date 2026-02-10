# Test plugin bằng curl (không cần OpenIG)

Mục tiêu: chạy Vault dev + plugin trong container, sau đó dùng `curl` gọi API để test `config/rotate/creds`.

## 1) Start containers

```bash
docker compose up --build
```

Vault sẽ listen tại `http://127.0.0.1:8200`.
Token dev mặc định: `root`.

## 2) Set env (tuỳ chọn)

```bash
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
```

## 3) Read config

```bash
curl -sS "$VAULT_ADDR/v1/legacy-cred/config" \
  -H "X-Vault-Token: $VAULT_TOKEN" | jq
```

## 4) Rotate (POST)

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/rotate/alice" \
  -H "X-Vault-Token: $VAULT_TOKEN" | jq
```

Bạn sẽ thấy `plaintext`, `hash`, `expires_at`.

## 5) Read creds (GET)

```bash
curl -sS "$VAULT_ADDR/v1/legacy-cred/creds/alice" \
  -H "X-Vault-Token: $VAULT_TOKEN" | jq
```

## 6) (Optional) Test webhook updater

Mock updater chạy tại `http://127.0.0.1:8080`.

Cấu hình plugin dùng webhook:

```bash
curl -sS -X POST "$VAULT_ADDR/v1/legacy-cred/config" \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -d '{
    "updater_type":"webhook",
    "webhook_url":"http://mock-updater:8080/update",
    "webhook_method":"POST"
  }' | jq
```

Sau đó rotate lại và xem log của container `mock-updater`.

## 7) Test SQL updater với DB tiêu chuẩn (Postgres/MySQL/MSSQL)

Repo đã có sẵn `docker-compose.yml` với 3 DB và init schema `users(username,password_hash)`.

Chạy test tự động (khuyến nghị):

```bash
chmod +x scripts/test-sql-updater.sh
./scripts/test-sql-updater.sh
```

Script sẽ:
- start `vault` + `postgres` + `mysql` + `mssql` (kèm init)
- cấu hình `legacy-cred/config` sang `updater_type=sql` theo từng DB
- gọi `rotate/alice` và verify `password_hash` trong DB khớp với `hash` trả về
