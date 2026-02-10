# Build & Test

## Prerequisites

- Go đã cài (trong môi trường hiện tại đã cài qua Homebrew)

## Test

```bash
go test ./...
```

## Build binary

```bash
go build -o vault-plugin-legacy-cred ./cmd/vault-plugin-legacy-cred
```

## Đăng ký plugin với Vault (ví dụ)

```bash
vault plugin register \
  -sha256=$(shasum -a 256 vault-plugin-legacy-cred | awk '{print $1}') \
  secret vault-plugin-legacy-cred

vault secrets enable -path=legacy-cred vault-plugin-legacy-cred
```

## Sanity test (manual)

```bash
vault write legacy-cred/config updater_type=noop
vault write legacy-cred/rotate/alice
vault read legacy-cred/creds/alice
```
