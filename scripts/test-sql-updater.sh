#!/usr/bin/env bash
set -euo pipefail

export VAULT_ADDR=${VAULT_ADDR:-http://127.0.0.1:8200}
export VAULT_TOKEN=${VAULT_TOKEN:-root}
export MSSQL_SA_PASSWORD=${MSSQL_SA_PASSWORD:-Passw0rd!}

require() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

require curl
require docker
require python3

wait_http_200() {
  local url="$1"
  local name="$2"
  local i=0
  until curl -sS -o /dev/null -w "%{http_code}" "$url" | grep -q '^200$'; do
    i=$((i+1))
    if [ "$i" -gt 200 ]; then
      echo "$name did not become ready: $url" >&2
      return 1
    fi
    sleep 0.2
  done
}

wait_vault_legacycred_ready() {
  local i=0
  while true; do
    i=$((i+1))
    # If the mount isn't enabled yet, Vault returns 404.
    # Once enabled, this endpoint should return 200.
    local code
    code=$(curl -sS -o /dev/null -w "%{http_code}" "$VAULT_ADDR/v1/legacy-cred/config" \
      -H "X-Vault-Token: $VAULT_TOKEN" || true)
    if [ "$code" = "200" ]; then
      return 0
    fi
    if [ "$i" -gt 200 ]; then
      echo "legacy-cred mount did not become ready (last HTTP $code)" >&2
      return 1
    fi
    sleep 0.2
  done
}

wait_postgres_ready() {
  local i=0
  until docker compose exec -T postgres pg_isready -U legacy -d legacy >/dev/null 2>&1; do
    i=$((i+1))
    if [ "$i" -gt 60 ]; then
      echo "Postgres did not become ready" >&2
      return 1
    fi
    sleep 1
  done
}

wait_mysql_ready() {
  local i=0
  until docker compose exec -T mysql mysqladmin ping -uroot -proot -h 127.0.0.1 --silent >/dev/null 2>&1; do
    i=$((i+1))
    if [ "$i" -gt 80 ]; then
      echo "MySQL did not become ready" >&2
      return 1
    fi
    sleep 1
  done
}

wait_mssql_init_done() {
  local cid
  cid=$(docker compose ps -aq mssql-init || true)
  if [ -z "$cid" ]; then
    echo "Could not find mssql-init container" >&2
    return 1
  fi
  local code
  code=$(docker wait "$cid")
  if [ "$code" != "0" ]; then
    echo "mssql-init failed with exit code $code" >&2
    docker logs "$cid" 2>/dev/null || true
    return 1
  fi
}

vault_config_sql() {
  local driver="$1"
  local dsn="$2"
  local query="$3"
  local style="$4"

  local out body code
  out=$(curl -sS -w $'\n%{http_code}' -X POST "$VAULT_ADDR/v1/legacy-cred/config" \
    -H "X-Vault-Token: $VAULT_TOKEN" \
    -d "{\"updater_type\":\"sql\",\"sql_driver\":\"$driver\",\"sql_dsn\":\"$dsn\",\"sql_update_query_named\":\"$query\",\"sql_placeholder_style\":\"$style\"}") || {
    echo "Vault config request failed" >&2
    return 1
  }
  body=${out%$'\n'*}
  code=${out##*$'\n'}
  if [ "$code" -lt 200 ] || [ "$code" -ge 300 ]; then
    echo "Vault config failed (HTTP $code)" >&2
    echo "$body" >&2
    return 1
  fi
  if [ -z "$body" ]; then
    echo "Vault config returned empty body" >&2
    return 1
  fi
}

vault_rotate_hash() {
  local subject="$1"
  local out body code
  out=$(curl -sS -w $'\n%{http_code}' -X POST "$VAULT_ADDR/v1/legacy-cred/rotate/$subject" -H "X-Vault-Token: $VAULT_TOKEN") || {
    echo "Vault rotate request failed" >&2
    return 1
  }
  body=${out%$'\n'*}
  code=${out##*$'\n'}
  if [ "$code" -lt 200 ] || [ "$code" -ge 300 ]; then
    echo "Vault rotate failed (HTTP $code)" >&2
    echo "$body" >&2
    return 1
  fi
  if [ -z "$body" ]; then
    echo "Vault rotate returned empty body" >&2
    return 1
  fi
  local tmp
  tmp=$(mktemp)
  printf '%s' "$body" >"$tmp"
  if [ ! -s "$tmp" ]; then
    echo "Vault rotate body file empty (HTTP $code)" >&2
    rm -f "$tmp"
    return 1
  fi
  if ! python3 -c 'import json,sys; j=json.load(open(sys.argv[1],"r",encoding="utf-8")); print(j["data"]["hash"])' "$tmp"; then
    echo "Vault rotate response is not valid JSON:" >&2
    head -c 500 "$tmp" >&2
    echo >&2
    rm -f "$tmp"
    return 1
  fi
  rm -f "$tmp"
}

postgres_reset_user() {
  local subject="$1"
  docker compose exec -T postgres psql -U legacy -d legacy -v ON_ERROR_STOP=1 \
    -c "UPDATE users SET password_hash='' WHERE username='${subject}';" >/dev/null
}

postgres_get_hash() {
  local subject="$1"
  docker compose exec -T postgres psql -U legacy -d legacy -tAc \
    "SELECT password_hash FROM users WHERE username='${subject}';" | tr -d '\r'
}

mysql_reset_user() {
  local subject="$1"
  docker compose exec -T mysql mysql -ulegacy -plegacy legacy \
    -e "UPDATE users SET password_hash='' WHERE username='${subject}';" >/dev/null
}

mysql_get_hash() {
  local subject="$1"
  docker compose exec -T mysql mysql -ulegacy -plegacy legacy -Nse \
    "SELECT password_hash FROM users WHERE username='${subject}';" | tr -d '\r'
}

mssql_reset_user() {
  local subject="$1"
  docker compose run --rm -T mssql-tools \
    "/opt/mssql-tools/bin/sqlcmd -S mssql -U sa -P \"$MSSQL_SA_PASSWORD\" -d legacy -Q \"SET NOCOUNT ON; UPDATE dbo.users SET password_hash='' WHERE username='${subject}';\" -b -C" \
    >/dev/null
}

mssql_get_hash() {
  local subject="$1"
  docker compose run --rm -T mssql-tools \
    "/opt/mssql-tools/bin/sqlcmd -S mssql -U sa -P \"$MSSQL_SA_PASSWORD\" -d legacy -Q \"SET NOCOUNT ON; SELECT password_hash FROM dbo.users WHERE username='${subject}';\" -h -1 -W -b -C" \
    | tr -d '\r'
}

main() {
  echo "[1/4] Starting Vault + standard DBs..."
  docker compose up -d --build vault postgres mysql mssql mssql-init >/dev/null

  echo "[2/4] Waiting for Vault HTTP..."
  wait_http_200 "$VAULT_ADDR/v1/sys/health" "Vault"

  echo "Waiting for legacy-cred mount..."
  wait_vault_legacycred_ready

  echo "Waiting for Postgres/MySQL ready and MSSQL init..."
  wait_postgres_ready
  wait_mysql_ready
  wait_mssql_init_done

  local subject="alice"

  echo "[3/4] Testing Postgres SQL updater..."
  postgres_reset_user "$subject"
  vault_config_sql \
    postgres \
    "postgres://legacy:legacy@postgres:5432/legacy?sslmode=disable" \
    "UPDATE users SET password_hash=:hash WHERE username=:subject" \
    dollar
  local hash_pg
  hash_pg=$(vault_rotate_hash "$subject")
  local db_pg
  db_pg=$(postgres_get_hash "$subject")
  if [ "$db_pg" != "$hash_pg" ]; then
    echo "Postgres verification failed" >&2
    echo "  expected: $hash_pg" >&2
    echo "  got:      $db_pg" >&2
    exit 1
  fi
  echo "  OK"

  echo "[4/4] Testing MySQL SQL updater..."
  mysql_reset_user "$subject"
  vault_config_sql \
    mysql \
    "legacy:legacy@tcp(mysql:3306)/legacy" \
    "UPDATE users SET password_hash=:hash WHERE username=:subject" \
    question
  local hash_my
  hash_my=$(vault_rotate_hash "$subject")
  local db_my
  db_my=$(mysql_get_hash "$subject")
  if [ "$db_my" != "$hash_my" ]; then
    echo "MySQL verification failed" >&2
    echo "  expected: $hash_my" >&2
    echo "  got:      $db_my" >&2
    exit 1
  fi
  echo "  OK"

  echo "[5/4] Testing MSSQL SQL updater..."
  mssql_reset_user "$subject"
  vault_config_sql \
    sqlserver \
    "sqlserver://sa:Passw0rd%21@mssql:1433?database=legacy&encrypt=disable" \
    "UPDATE dbo.users SET password_hash=:hash WHERE username=:subject" \
    at
  local hash_ms
  hash_ms=$(vault_rotate_hash "$subject")
  local db_ms
  db_ms=$(mssql_get_hash "$subject")
  if [ "$db_ms" != "$hash_ms" ]; then
    echo "MSSQL verification failed" >&2
    echo "  expected: $hash_ms" >&2
    echo "  got:      $db_ms" >&2
    exit 1
  fi
  echo "  OK"

  echo "All SQL updater tests passed."
}

main "$@"
