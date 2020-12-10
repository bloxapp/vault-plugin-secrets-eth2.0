#!/bin/sh

mkdir -p /data/keys
vault operator init -status
if [[ $? -eq 2 ]]; then
  vault operator init \
      -key-shares=1 \
      -key-threshold=1 \
      -recovery-shares=1 \
      -recovery-threshold=1 \
      -format=json > /tmp/vault.init 2>&1

  cat /tmp/vault.init | jq -r '.root_token' > /data/keys/vault.root.token && chmod 0677 /data/keys/vault.root.token
  cat /tmp/vault.init | jq -r '.unseal_keys_b64[]' > /data/keys/vault.unseal.token && chmod 0677 /data/keys/vault.unseal.token
  rm -f /tmp/vault.*
fi 
