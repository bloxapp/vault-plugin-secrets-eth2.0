#!/bin/bash

function banner() {
  echo "+----------------------------------------------------------------------------------+"
  printf "| %-80s |\n" "$(date)"
  echo "|                                                                                  |"
  printf "| %-80s |\n" "$@"
  echo "+----------------------------------------------------------------------------------+"
}

function register_plugin() {
  banner "Registering plugin ethsign"
  set -eo pipefail
  SHASUM256=$(sha256sum "/vault/plugins/ethsign" | cut -d' ' -f1)
  vault plugin register \
    -sha256="${SHASUM256}" \
    -args="--log-format=${LOG_FORMAT}" \
    -args="--log-dsn=${LOG_DSN}" \
    -args="--log-levels=${LOG_LEVELS}" \
    secret ethsign
}

function enable_plugin() {
  banner "Re-enabling plugin $1"
  vault secrets disable ethereum/"$1"
  vault secrets enable \
      -path=ethereum/"$1" \
      -description="Eth Signing Wallet - $1 network" \
      -plugin-name=ethsign plugin
}

function configure_network() {
  banner "Configuring network $1"
  vault write ethereum/"$1"/config \
      network="$1"
}

function reload_backend() {
  banner "Reloading backend"
  curl --insecure --header "X-Vault-Token: $VAULT_TOKEN" \
     --request PUT \
     --data '{"plugin": "ethsign"}' \
     "${VAULT_SERVER_SCHEMA:-http}"://127.0.0.1:8200/v1/sys/plugins/reload/backend
}

function health_check() {
  banner "Health check $1"
  curl --insecure \
       --header "X-Vault-Token: $VAULT_TOKEN" \
       --request GET \
       --fail \
       "${VAULT_SERVER_SCHEMA:-http}"://127.0.0.1:8200/v1/ethereum/"$1"/config
}

function write_policy() {
  banner "Writing $1 policy"
  vault policy write "$1" vault/policies/"$1"-policy.hcl
}

banner "Configuring ethsign plugin at $VAULT_ADDR..."
register_plugin

enable_plugin pyrmont
enable_plugin mainnet

configure_network pyrmont
configure_network mainnet

reload_backend

health_check pyrmont
health_check mainnet
