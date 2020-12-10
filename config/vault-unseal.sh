#!/bin/sh

vault status
if [[ $? -eq 2 ]]; then
  vault operator unseal "$(cat /data/keys/vault.unseal.token)" > /dev/null 2>&1
  vault login "$(cat /data/keys/vault.root.token)" > /dev/null 2>&1
fi 
