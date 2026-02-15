#!/usr/bin/env bash
set -euo pipefail

# Migrates external (http/https) content section image URLs into locally served
# /images/content/* files by downloading and updating DB rows.
#
# Requirements:
# - docker containers: sangehassan-db, sangehassan-backend
# - curl available on host

DB_CONTAINER="${DB_CONTAINER:-sangehassan-db}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-sangehassan-backend}"
DB_USER="${DB_USER:-sangehassan}"
DB_NAME="${DB_NAME:-sangehassan}"

rows="$(
  docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -tA -F $'\t' \
    -c "select id, image_url from content_section_images where image_url like 'http%';"
)"

if [[ -z "${rows}" ]]; then
  echo "No external content_section_images rows found."
  exit 0
fi

docker exec "$BACKEND_CONTAINER" sh -lc 'mkdir -p /app/storage/images/content'

TMP_ROOT="${TMPDIR:-$(pwd)/.tmp}"
mkdir -p "${TMP_ROOT}"

while IFS=$'\t' read -r id url; do
  [[ -z "${id}" || -z "${url}" ]] && continue
  out_name="pexels-${id}.jpg"
  out_path="/images/content/${out_name}"

  tmp="$(mktemp -p "${TMP_ROOT}")"
  echo "Downloading id=${id} ..."
  if ! curl -fsSL -L "${url}" -o "${tmp}"; then
    echo "  WARN: download failed for id=${id}. Using local placeholder instead."
    # Round-robin placeholders to keep some visual variety.
    case $((id % 3)) in
      0) cp "back/storage/images/granite.jpg" "${tmp}" ;;
      1) cp "back/storage/images/marble.jpg" "${tmp}" ;;
      2) cp "back/storage/images/travertine.jpg" "${tmp}" ;;
    esac
  fi

  docker cp "${tmp}" "${BACKEND_CONTAINER}:/app/storage/images/content/${out_name}"
  docker exec "${BACKEND_CONTAINER}" sh -lc "chmod 0644 /app/storage/images/content/${out_name} || true"
  rm -f "${tmp}"

  docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -q \
    -c "update content_section_images set image_url='${out_path}' where id=${id};"
done <<< "${rows}"

echo "Done."
