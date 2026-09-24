#!/usr/bin/env bash
# Creates .env from .env.example, replacing each placeholder with a unique value:
#   __HEX__     -> 48 hex characters
#   __B64_32__  -> 32 random bytes, base64-encoded
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ -e .env ]]; then
  echo ".env already exists; refusing to overwrite it." >&2
  exit 1
fi

umask 077

while IFS= read -r line || [[ -n "$line" ]]; do
  while [[ "$line" == *__HEX__* ]]; do
    line="${line/__HEX__/$(openssl rand -hex 24)}"
  done

  while [[ "$line" == *__B64_32__* ]]; do
    line="${line/__B64_32__/$(openssl rand -base64 32)}"
  done

  printf '%s\n' "$line"
done < .env.example > .env

echo "Wrote $(pwd)/.env. Review DEMO_DOMAIN and the optional integrations before starting."
