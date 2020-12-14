#!/bin/bash

set -eo pipefail

export SHASUM256=$(sha256sum "/vault/plugins/ethsign" | cut -d' ' -f1)

# Register plugin
vault plugin register \
    -sha256=${SHASUM256} \
    -args="--log-format=${LOG_FORMAT}" \
    -args="--log-dsn=${LOG_DSN}" \
    -args="--log-levels=${LOG_LEVELS}" \
    secret ethsign

# Enable networks
echo "Enabling plugin..."
vault secrets disable ethereum/pyrmont
vault secrets enable \
    -path=ethereum/pyrmont \
    -description="Eth Signing Wallet - Pyrmont Test network" \
    -plugin-name=ethsign plugin
echo "Enabled plugin"

echo "Enabling plugin..."
vault secrets disable ethereum/mainnet
vault secrets enable \
    -path=ethereum/mainnet \
    -description="Eth Signing Wallet - MainNet" \
    -plugin-name=ethsign plugin
echo "Enabled plugin"

# Configuring networks
echo "Configuring Pyrmont Test network..."
vault write ethereum/pyrmont/config \
    network="pyrmont"
echo "Configured Pyrmont Test network"

echo "Configuring MainNet network..."
vault write ethereum/mainnet/config \
    network="mainnet"
echo "Configured MainNet network"

# Reload plugin
curl --insecure --header "X-Vault-Token: $(cat /data/keys/vault.root.token)" \
     --request PUT \
     --data '{"plugin": "ethsign"}' \
     ${VAULT_SERVER_SCHEMA:-http}://127.0.0.1:8200/v1/sys/plugins/reload/backend

# Make sure everything works well
curl --insecure \
     --header "X-Vault-Token: $(cat /data/keys/vault.root.token)" \
     --request GET \
     --fail \
     ${VAULT_SERVER_SCHEMA:-http}://127.0.0.1:8200/v1/ethereum/pyrmont/config

curl --insecure \
     --header "X-Vault-Token: $(cat /data/keys/vault.root.token)" \
     --request GET \
     --fail \
     ${VAULT_SERVER_SCHEMA:-http}://127.0.0.1:8200/v1/ethereum/mainnet/config
