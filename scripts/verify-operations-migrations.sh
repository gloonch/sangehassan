#!/bin/sh
set -eu

repo_dir="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
expected_file="$(mktemp)"
actual_file="$(mktemp)"
trap 'rm -f "$expected_file" "$actual_file"' EXIT HUP INT TERM

cat >"$expected_file" <<'CHECKSUMS'
d84b2d23d4476a1cbf6edbdfd84c30ac71957d5452fefee9cee64628c1a2e8aa  deploy/postgres/init/014_operations_phase1.sql
d2b390abfbfb821927dcd1e6a388b82df6b220564179a8568e157471b0e48e3d  deploy/postgres/init/015_operations_phase2.sql
222a0ebd88fc8e7f1acdd1ab39daf026ad4aa29efd0ea7bbe9ade2aa84405af8  deploy/postgres/init/016_operations_phase3.sql
5ca954f96e4b09ed9f071d64103a66c68c0fafe827d934f5a552cb5d4082dacb  deploy/postgres/init/017_finance_notifications_documents_reporting.sql
6fc176841510325ab5e7961887fed86059ae7df1ab47dc25a33059e9bd0b4a5b  deploy/postgres/init/018_supplier_purchase_quality_installation.sql
CHECKSUMS

cd "$repo_dir"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum deploy/postgres/init/014_operations_phase1.sql deploy/postgres/init/015_operations_phase2.sql deploy/postgres/init/016_operations_phase3.sql deploy/postgres/init/017_finance_notifications_documents_reporting.sql deploy/postgres/init/018_supplier_purchase_quality_installation.sql >"$actual_file"
else
  shasum -a 256 deploy/postgres/init/014_operations_phase1.sql deploy/postgres/init/015_operations_phase2.sql deploy/postgres/init/016_operations_phase3.sql deploy/postgres/init/017_finance_notifications_documents_reporting.sql deploy/postgres/init/018_supplier_purchase_quality_installation.sql >"$actual_file"
fi

if ! diff -u "$expected_file" "$actual_file"; then
  echo "Protected migration checksum mismatch." >&2
  exit 1
fi
echo "Protected migration checksums are valid."
