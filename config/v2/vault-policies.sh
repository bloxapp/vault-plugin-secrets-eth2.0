#!/bin/bash

function banner() {
  echo "+----------------------------------------------------------------------------------+"
  printf "| %-80s |\n" "$(date)"
  echo "|                                                                                  |"
  printf "| %-80s |\n" "$@"
  echo "+----------------------------------------------------------------------------------+"
}

function write_policy() {
  banner "Writing $1 policy"
  vault policy write "$1" vault/policies/"$1"-policy.hcl
  echo "$1".token="$(vault token create -orphan=true -policy="$1" -field=token)"
}

write_policy signer
write_policy admin
