#!/bin/sh
set -eu

mode="${1:-}"
backup_path="${2:-}"

if [ "$mode" != "verify" ] && [ "$mode" != "migrate" ]; then
  echo "usage: $0 verify|migrate /path/to/backup.dump" >&2
  exit 2
fi
if [ -z "$backup_path" ] || [ ! -f "$backup_path" ] || [ ! -s "$backup_path" ]; then
  echo "backup file is missing or empty: $backup_path" >&2
  exit 1
fi

checksum="$(sha256sum "$backup_path" | awk '{print $1}')"

verify_with_pg_restore() {
  archive="$1"
  if command -v pg_restore >/dev/null 2>&1; then
    pg_restore --list "$archive" >/dev/null
    return
  fi
  if ! command -v docker >/dev/null 2>&1; then
    echo "pg_restore or Docker is required to validate the backup" >&2
    exit 1
  fi
  container="${POSTGRES_CONTAINER:-erp-empi-postgres-1}"
  container_archive="/tmp/empi-predeploy-backup.dump"
  docker cp "$archive" "$container:$container_archive" >/dev/null
  trap 'docker exec "$container" rm -f "$container_archive" >/dev/null 2>&1 || true' EXIT HUP INT TERM
  docker exec "$container" pg_restore --list "$container_archive" >/dev/null
}

verify_with_pg_restore "$backup_path"
echo "backup verified"
echo "sha256=$checksum"

if [ "$mode" = "verify" ]; then
  exit 0
fi

if [ -z "${BACKUP_SHA256:-}" ] || [ "$BACKUP_SHA256" != "$checksum" ]; then
  echo "BACKUP_SHA256 is required and must match the verified backup" >&2
  exit 1
fi
if [ -z "${DB_WRITE_DSN:-}" ]; then
  echo "DB_WRITE_DSN is required" >&2
  exit 1
fi
if [ "${CONFIRM_PRODUCTION_MIGRATION:-}" != "migrate-legacy-production" ]; then
  echo "set CONFIRM_PRODUCTION_MIGRATION=migrate-legacy-production to continue" >&2
  exit 1
fi

GOCACHE="${GOCACHE:-/tmp/erp-empi-go-cache}" go run ./cmd/migrate
