#!/bin/sh
set -eu

repo_dir="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
container_name="sangehassan-migration-smoke-$$"
database_name="sangehassan_smoke"
database_user="smoke"
database_password="local-smoke-only"

cleanup() {
  docker rm -f "$container_name" >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM

docker run --detach --name "$container_name" \
  -e POSTGRES_USER="$database_user" \
  -e POSTGRES_PASSWORD="$database_password" \
  -e POSTGRES_DB="$database_name" \
  postgres:16-alpine >/dev/null

attempt=0
until docker exec "$container_name" pg_isready -U "$database_user" -d "$database_name" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    echo "PostgreSQL 16 did not become ready." >&2
    exit 1
  fi
  sleep 1
done

apply_sql() {
  sql_file="$1"
  echo "Applying $(basename "$sql_file")"
  docker exec -i "$container_name" psql -v ON_ERROR_STOP=1 -U "$database_user" -d "$database_name" <"$sql_file" >/dev/null
}

for sql_file in "$repo_dir"/deploy/postgres/init/*.sql; do
  apply_sql "$sql_file"
done

for sql_file in \
  "$repo_dir/deploy/postgres/init/014_operations_phase1.sql" \
  "$repo_dir/deploy/postgres/init/015_operations_phase2.sql" \
  "$repo_dir/deploy/postgres/init/016_operations_phase3.sql" \
  "$repo_dir/deploy/postgres/init/017_finance_notifications_documents_reporting.sql" \
  "$repo_dir/deploy/postgres/init/018_supplier_purchase_quality_installation.sql" \
  "$repo_dir/deploy/postgres/init/019_application_settings_diagnostics_indexes.sql" \
  "$repo_dir/deploy/postgres/init/020_product_display_order.sql"; do
  apply_sql "$sql_file"
done

migration_version="$(docker exec "$container_name" psql -At -U "$database_user" -d "$database_name" -c 'SELECT COALESCE(MAX(version),0) FROM schema_migrations')"
if [ "$migration_version" != "20" ]; then
  echo "Expected migration version 20, got $migration_version." >&2
  exit 1
fi

echo "PostgreSQL 16 migration smoke test passed."
