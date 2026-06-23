#!/bin/sh
set -eu

: "${SUPPORT_ADMIN_DB_PASSWORD:?SUPPORT_ADMIN_DB_PASSWORD is required}"

psql -v ON_ERROR_STOP=1 \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --set=support_password="$SUPPORT_ADMIN_DB_PASSWORD" <<'EOSQL'
SELECT format('CREATE ROLE support_admin LOGIN PASSWORD %L', :'support_password')
WHERE NOT EXISTS (
  SELECT FROM pg_roles WHERE rolname = 'support_admin'
)\gexec

SELECT 'CREATE DATABASE support_admin OWNER support_admin'
WHERE NOT EXISTS (
  SELECT FROM pg_database WHERE datname = 'support_admin'
)\gexec
EOSQL
