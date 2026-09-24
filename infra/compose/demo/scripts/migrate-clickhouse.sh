#!/bin/sh
# Applies infra/clickhouse/migrations to the demo ClickHouse node.
# Mirrors the migrate-clickhouse job in shahrah-gitops with replicatedDatabase enabled.
set -eu

client() {
  clickhouse-client \
    --host "${CLICKHOUSE_HOST}" \
    --user "${CLICKHOUSE_USER}" \
    --password "${CLICKHOUSE_PASSWORD}" \
    "$@"
}

client --database default \
  --query "CREATE DATABASE IF NOT EXISTS \`${CLICKHOUSE_DATABASE}\` ENGINE = Replicated('/clickhouse/databases/${CLICKHOUSE_DATABASE}', '{shard}', '{replica}')"

for migration in $(find -L /migrations -maxdepth 1 -type f -name '*.sql' | sort); do
  echo "Applying ${migration}"
  # Strip a UTF-8 BOM if present; clickhouse-client rejects it.
  sed '1s/^\xEF\xBB\xBF//' "${migration}" | client --database "${CLICKHOUSE_DATABASE}" --multiquery
done

echo "ClickHouse migrations applied."
