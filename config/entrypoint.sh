#!/bin/bash

set -eo pipefail

mkdir -p /data

VAULT_SERVER_CONFIG=/vault/config/vault-config.json
VAULT_SERVER_SCHEMA=http

# Generate SSL certificates if external IP address is provided
if [ "$VAULT_EXTERNAL_ADDRESS" != "" ]; then
  export VAULT_CACERT=/vault/config/ca.pem
  VAULT_SERVER_CONFIG=/vault/config/vault-config-tls.json
  VAULT_SERVER_SCHEMA=https

  /vault/config/vault-tls.sh
fi

export VAULT_SERVER_SCHEMA=$VAULT_SERVER_SCHEMA
export VAULT_ADDR=$VAULT_SERVER_SCHEMA://127.0.0.1:8200
export VAULT_API_ADDR=$VAULT_SERVER_SCHEMA://127.0.0.1:8200

# Start vault server
touch /data/logs
vault server -config=$VAULT_SERVER_CONFIG -log-level=debug > /data/logs &
echo "started vault server"

sleep 5
if [ "$UNSEAL" = "true" ]; then
  /vault/config/vault-init.sh
  /vault/config/vault-unseal.sh
  /vault/config/vault-plugin.sh
fi

sleep 356000d
