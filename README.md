# Vault Legacy Credential Rotator (Plugin)

This repository contains a Vault **secrets engine plugin** that helps integrate legacy applications into a centralized SSO flow (e.g. Keycloak + OpenIG) when you **cannot modify legacy app code**.

Detailed documentation (Vietnamese) is in [docs/README.md](docs/README.md).

## What it does

- Generate a random plaintext password for a `subject` (username/email)
- Hash the plaintext using a configured algorithm (default: `bcrypt`)
- Update the legacy credential store (via `webhook` or `sql` updater)
- Store plaintext in Vault for a limited time (default TTL: **3 hours**) so OpenIG can fetch it and perform legacy login

> Note: Storing plaintext is generally discouraged. This plugin exists specifically for “credential translation” for legacy systems. Use strict Vault policies and short TTLs.

## Build

```bash
go build -o vault-plugin-legacy-cred ./cmd/vault-plugin-legacy-cred
```

## Git: end-of-day sync

If you want a single command to stage/commit/push all changes to GitHub at the end of your workday:

```bash
chmod +x scripts/*.sh
./scripts/eod-push.sh "chore: end of day sync"
```

See [scripts/README.md](scripts/README.md) for details.

## Vault: register + enable (example)

```bash
# Register the plugin binary
vault plugin register \
  -sha256=$(shasum -a 256 vault-plugin-legacy-cred | awk '{print $1}') \
  secret vault-plugin-legacy-cred

# Enable under a mount path
vault secrets enable -path=legacy-cred vault-plugin-legacy-cred
```

## Configure

```bash
vault write legacy-cred/config \
  default_ttl=10800 \
  default_hash_type=bcrypt \
  bcrypt_cost=12 \
  updater_type=noop
```

### Webhook updater (recommended for multi-DB trials)

```bash
vault write legacy-cred/config \
  updater_type=webhook \
  webhook_url="https://legacy-updater.example.com/update" \
  webhook_method=POST \
  webhook_headers='{"Authorization":"Bearer ..."}'
```

The webhook receives JSON:

```json
{"subject":"user@example.com","hash":"...","hash_type":"bcrypt","version":1,"updated_at":"2026-02-10T00:00:00Z"}
```

### SQL updater (direct DB)

```bash
vault write legacy-cred/config \
  updater_type=sql \
  sql_driver=postgres \
  sql_dsn="postgres://user:pass@host:5432/db?sslmode=disable" \
  sql_update_query_named='UPDATE users SET password_hash=:hash WHERE username=:subject' \
  sql_placeholder_style=dollar
```

Drivers included:
- Postgres (`pgx` stdlib): `postgres`
- MySQL: `mysql`
- MS SQL Server: `sqlserver`

## Rotate

```bash
vault write legacy-cred/rotate/user@example.com
```

Returns `plaintext`, `hash`, `expires_at`.

## Read creds (OpenIG)

```bash
vault read legacy-cred/creds/user@example.com
```

Tip: Use Vault response wrapping (`X-Vault-Wrap-TTL`) from OpenIG to reduce plaintext exposure.

## Security notes

- Create a dedicated policy for OpenIG to only `read` `legacy-cred/creds/*`
- Create a separate policy for rotation jobs to `update` `legacy-cred/rotate/*`
- Keep TTL short (default 3h) and consider `one_time_read=true` when feasible
