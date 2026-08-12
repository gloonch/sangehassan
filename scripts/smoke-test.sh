#!/bin/sh
set -eu

: "${SMOKE_BASE_URL:?Set SMOKE_BASE_URL, for example https://sangehassan.com}"

smoke_tmp_dir="$(mktemp -d)"
trap 'rm -rf "$smoke_tmp_dir"' EXIT HUP INT TERM

request() {
  endpoint="$1"
  expected="$2"
  code="$(curl --silent --show-error --output "$smoke_tmp_dir/response.json" --write-out '%{http_code}' "$SMOKE_BASE_URL$endpoint")"
  if [ "$code" != "$expected" ]; then
    echo "Smoke check failed: $endpoint returned HTTP $code" >&2
    exit 1
  fi
}

request /health 200
request /ready 200
request /api/v1/version 200

if [ -n "${SMOKE_INTERNAL_PHONE:-}" ] || [ -n "${SMOKE_INTERNAL_PASSWORD:-}" ]; then
  : "${SMOKE_INTERNAL_PHONE:?Both SMOKE_INTERNAL_PHONE and SMOKE_INTERNAL_PASSWORD are required}"
  : "${SMOKE_INTERNAL_PASSWORD:?Both SMOKE_INTERNAL_PHONE and SMOKE_INTERNAL_PASSWORD are required}"
  login_code="$(curl --silent --show-error --cookie-jar "$smoke_tmp_dir/cookies" --output "$smoke_tmp_dir/login.json" --write-out '%{http_code}' \
    -H 'Content-Type: application/json' \
    --data "{\"phone\":\"$SMOKE_INTERNAL_PHONE\",\"password\":\"$SMOKE_INTERNAL_PASSWORD\"}" \
    "$SMOKE_BASE_URL/api/v1/auth/internal/login")"
  if [ "$login_code" != 200 ]; then
    echo "Internal login smoke check failed with HTTP $login_code" >&2
    exit 1
  fi
  session_code="$(curl --silent --show-error --cookie "$smoke_tmp_dir/cookies" --output "$smoke_tmp_dir/session.json" --write-out '%{http_code}' "$SMOKE_BASE_URL/api/v1/operations/me")"
  if [ "$session_code" != 200 ]; then
    echo "Authenticated session smoke check failed with HTTP $session_code" >&2
    exit 1
  fi
fi

echo "Smoke checks passed."
