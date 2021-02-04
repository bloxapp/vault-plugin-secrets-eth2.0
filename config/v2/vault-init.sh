#!/bin/bash

function banner() {
  echo "+----------------------------------------------------------------------------------+"
  printf "| %-80s |\n" "$(date)"
  echo "|                                                                                  |"
  printf "| %-80s |\n" "$@"
  echo "+----------------------------------------------------------------------------------+"
}

function init() {
  banner "Initiating..."
  SECRETS=$(vault operator init -key-shares=1 -key-threshold=1 -format=json | jq .)
}

function unseal() {
  banner "Unsealing $VAULT_ADDR..."
  UNSEAL=$(echo "$SECRETS" | jq -r '.unseal_keys_hex[0]')
  vault operator unseal "$UNSEAL"
}

function authenticate() {
  banner "Authenticating to $VAULT_ADDR as root"
  ROOT=$(echo "$SECRETS" | jq -r .root_token)
  banner "ROOT $ROOT"
  export VAULT_TOKEN=$ROOT
}

function revoke_root_token() {
  banner "Unsetting VAULT_TOKEN"
  vault token revoke -self
  unset VAULT_TOKEN
}

function status() {
  vault status
}

vault operator init -status
if [[ $? -eq 2 ]]; then
  init
  unseal
  authenticate
  /bin/bash /vault/config/v2/vault-plugin.sh
  /bin/bash /vault/config/v2/vault-policies.sh
  revoke_root_token
  status
else
  unseal
  status
fi
