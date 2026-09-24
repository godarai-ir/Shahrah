#!/usr/bin/env bash
# Nightly demo backup: Postgres databases and support attachments.
# ClickHouse (usage/trace analytics) is intentionally not backed up.
#
# Cron example (03:30 every night):
#   30 3 * * * /opt/shahrah/infra/compose/demo/scripts/backup.sh >> /var/log/shahrah-backup.log 2>&1
set -euo pipefail

cd "$(dirname "$0")/.."

set -a
# shellcheck disable=SC1091
source .env
set +a

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="backups/${stamp}"
mkdir -p "${backup_dir}"

compose() {
  docker compose --env-file .env "$@"
}

for db in sharah_db model_discovery support_admin; do
  echo "Dumping ${db}"
  compose exec -T postgres pg_dump -U postgres --format=custom "${db}" > "${backup_dir}/${db}.dump"
done

echo "Copying support attachments"
mkdir -p "${backup_dir}/support-files"
compose run --rm --no-deps -T -v "$(pwd)/${backup_dir}/support-files:/out" --entrypoint sh minio-init -c \
  'mc alias set local http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null &&
   mc mirror --quiet local/godarai-support /out'

if [[ -n "${BACKUP_S3_ENDPOINT:-}" && -n "${BACKUP_S3_BUCKET:-}" ]]; then
  echo "Uploading to ${BACKUP_S3_ENDPOINT}/${BACKUP_S3_BUCKET}/${stamp}"
  docker run --rm -v "$(pwd)/${backup_dir}:/backup:ro" \
    -e MC_HOST_remote="https://${BACKUP_S3_ACCESS_KEY}:${BACKUP_S3_SECRET_KEY}@${BACKUP_S3_ENDPOINT#https://}" \
    minio/mc:latest cp --recursive /backup/ "remote/${BACKUP_S3_BUCKET}/${stamp}/"
fi

find backups -mindepth 1 -maxdepth 1 -type d -mtime "+${BACKUP_LOCAL_RETENTION_DAYS:-7}" -exec rm -rf {} +

echo "Backup ${stamp} complete."
