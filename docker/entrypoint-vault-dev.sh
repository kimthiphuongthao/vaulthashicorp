#!/usr/bin/env sh
set -eu

export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="${VAULT_DEV_ROOT_TOKEN_ID:-root}"

# Start Vault dev server with plugin directory
vault server \
  -dev \
  -dev-root-token-id="$VAULT_TOKEN" \
  -dev-listen-address="0.0.0.0:8200" \
  -dev-plugin-dir="/vault/plugins" &

VAULT_PID=$!

# Wait until ready
i=0
until vault status >/dev/null 2>&1; do
  i=$((i+1))
  if [ $i -gt 200 ]; then
    echo "Vault did not become ready in time" >&2
    kill $VAULT_PID || true
    exit 1
  fi
  sleep 0.1
done

# Compute sha256
if command -v sha256sum >/dev/null 2>&1; then
  PLUGIN_SHA=$(sha256sum /vault/plugins/vault-plugin-legacy-cred | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  PLUGIN_SHA=$(shasum -a 256 /vault/plugins/vault-plugin-legacy-cred | awk '{print $1}')
else
  PLUGIN_SHA=$(openssl dgst -sha256 /vault/plugins/vault-plugin-legacy-cred | awk '{print $2}')
fi

# Register & enable secrets engine
vault plugin register -sha256="$PLUGIN_SHA" secret vault-plugin-legacy-cred >/dev/null
vault secrets enable -path=legacy-cred vault-plugin-legacy-cred >/dev/null

# Default config for curl tests
vault write legacy-cred/config \
  updater_type=noop \
  default_ttl=10800 \
  default_hash_type=bcrypt \
  bcrypt_cost=12 >/dev/null

echo "Vault dev is ready on :8200 (token: $VAULT_TOKEN, mount: legacy-cred/)"

echo "Try from host: export VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=$VAULT_TOKEN"

echo "curl rotate: curl -sS -X POST \"$VAULT_ADDR/v1/legacy-cred/rotate/alice\" -H \"X-Vault-Token: $VAULT_TOKEN\" | jq"

wait $VAULT_PID
