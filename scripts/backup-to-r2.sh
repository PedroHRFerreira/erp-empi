#!/bin/sh
set -eu

require_value() {
  name="$1"
  value="$2"
  if [ -z "$value" ]; then
    echo "required environment variable is missing: $name" >&2
    exit 1
  fi
}

require_value DATABASE_URL "${DATABASE_URL:-}"
require_value R2_ENDPOINT "${R2_ENDPOINT:-}"
require_value R2_BUCKET "${R2_BUCKET:-}"
require_value AWS_ACCESS_KEY_ID "${AWS_ACCESS_KEY_ID:-}"
require_value AWS_SECRET_ACCESS_KEY "${AWS_SECRET_ACCESS_KEY:-}"

backup_prefix="${BACKUP_PREFIX:-erp-empi}"
timestamp="$(date -u +%Y-%m-%dT%H-%M-%SZ)"
year="$(date -u +%Y)"
month="$(date -u +%m)"
day="$(date -u +%d)"
workdir="$(mktemp -d /tmp/empi-backup.XXXXXX)"
archive="$workdir/$backup_prefix-$timestamp.dump"
checksum_file="$archive.sha256"

cleanup() {
  rm -rf "$workdir"
}
trap cleanup EXIT HUP INT TERM

pg_dump --dbname="$DATABASE_URL" --format=custom --no-owner --no-acl --file="$archive"
pg_restore --list "$archive" >/dev/null
checksum="$(sha256sum "$archive" | awk '{print $1}')"
printf '%s  %s\n' "$checksum" "$(basename "$archive")" >"$checksum_file"

daily_key="$backup_prefix/daily/$year/$month/$(basename "$archive")"
daily_checksum_key="$daily_key.sha256"

# R2 uses its own encryption at rest. These variables avoid metadata-service
# lookups and optional checksum headers that are not required by its S3 API.
export AWS_EC2_METADATA_DISABLED=true
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-auto}"
export AWS_REQUEST_CHECKSUM_CALCULATION=WHEN_REQUIRED
export AWS_RESPONSE_CHECKSUM_VALIDATION=WHEN_REQUIRED

aws --endpoint-url "$R2_ENDPOINT" s3 cp "$archive" "s3://$R2_BUCKET/$daily_key" --only-show-errors
aws --endpoint-url "$R2_ENDPOINT" s3 cp "$checksum_file" "s3://$R2_BUCKET/$daily_checksum_key" --only-show-errors

local_size="$(wc -c <"$archive" | tr -d ' ')"
remote_size="$(aws --endpoint-url "$R2_ENDPOINT" s3api head-object --bucket "$R2_BUCKET" --key "$daily_key" --query ContentLength --output text)"
if [ "$local_size" != "$remote_size" ]; then
  echo "uploaded backup size mismatch: local=$local_size remote=$remote_size" >&2
  exit 1
fi

if [ "$day" = "01" ]; then
  monthly_key="$backup_prefix/monthly/$year/$year-$month.dump"
  monthly_checksum_key="$monthly_key.sha256"
  monthly_checksum_file="$workdir/$year-$month.dump.sha256"
  printf '%s  %s\n' "$checksum" "$year-$month.dump" >"$monthly_checksum_file"
  aws --endpoint-url "$R2_ENDPOINT" s3 cp "$archive" "s3://$R2_BUCKET/$monthly_key" --only-show-errors
  aws --endpoint-url "$R2_ENDPOINT" s3 cp "$monthly_checksum_file" "s3://$R2_BUCKET/$monthly_checksum_key" --only-show-errors
fi

echo "backup uploaded and verified: s3://$R2_BUCKET/$daily_key"
echo "sha256=$checksum"
