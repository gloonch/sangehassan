#!/usr/bin/env bash
set -Eeuo pipefail

SITE_URL="${SITE_URL:-https://sangehassan.com}"
PROJECT_ROOT="${PROJECT_ROOT:-/opt/sangehassan}"
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_ROOT/deploy/docker-compose-prod-com.yml}"
STATE_DIR="${STATE_DIR:-/var/lib/sangehassan}"
STATE_FILE="${STATE_FILE:-$STATE_DIR/blog-prerender.fingerprint}"
LOCK_FILE="${LOCK_FILE:-/run/sangehassan-blog-prerender.lock}"
LOCALES="${LOCALES:-fa en ar}"
FORCE="${FORCE:-0}"
WEBSITE_IMAGE_REF="${WEBSITE_IMAGE_REF:-deploy-website:latest}"
DB_CONTAINER="${DB_CONTAINER:-sangehassan-db}"
DB_USER="${DB_USER:-sangehassan}"
DB_NAME="${DB_NAME:-sangehassan}"

log() {
  printf '[%s] %s\n' "$(date -Is)" "$*"
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    log "missing required command: $1"
    exit 1
  }
}

absolute_url() {
  local value="$1"
  case "$value" in
    http://*|https://*) printf '%s\n' "$value" ;;
    /*) printf '%s%s\n' "$SITE_URL" "$value" ;;
    "") printf '%s\n' "" ;;
    *) printf '%s/%s\n' "$SITE_URL" "$value" ;;
  esac
}

fetch_public_blogs() {
  : >"$BLOGS_TSV"

  for locale in $LOCALES; do
    local offset=0
    local total=0
    while :; do
      local response="$TMP_DIR/blogs-$locale-$offset.json"
      local url="$SITE_URL/api/blogs?locale=$locale&limit=100&offset=$offset"
      log "fetching $url"
      curl -fsS --retry 2 --retry-delay 3 --max-time 30 "$url" -o "$response"
      jq -r --arg locale "$locale" '
        (.data.items // [])
        | .[]
        | select((.slug // "") != "")
        | [
            $locale,
            (.slug // ""),
            (.updated_at // ""),
            (.published_at // ""),
            (.robots // "index,follow"),
            (.canonical_url // "")
          ]
        | @tsv
      ' "$response" >>"$BLOGS_TSV"
      total="$(jq -r '.data.total // 0' "$response")"
      offset=$((offset + 100))
      [ "$offset" -lt "$total" ] || break
    done
  done

  # Preserve API order because the blog listing page is prerendered in this order.
  awk '!seen[$0]++' "$BLOGS_TSV" >"$BLOGS_SORTED_TSV"
  sha256sum "$BLOGS_SORTED_TSV" | awk '{print $1}' >"$FINGERPRINT_FILE"
}

promote_due_scheduled_blogs() {
  local promoted

  promoted="$(docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -tAc "
    WITH due AS (
      UPDATE blogs
      SET status = 'published',
          published_at = COALESCE(scheduled_at, published_at, NOW()),
          scheduled_at = NULL,
          updated_at = NOW()
      WHERE status = 'scheduled'
        AND scheduled_at IS NOT NULL
        AND scheduled_at <= NOW()
      RETURNING id
    )
    SELECT COUNT(*) FROM due;
  ")"
  promoted="$(printf '%s' "$promoted" | tr -d '[:space:]')"

  if [ "${promoted:-0}" != "0" ]; then
    log "promoted $promoted scheduled blog(s) to published"
  fi
}

verify_blog_routes() {
  local base_url="$1"
  local failures=0
  local index=0

  while IFS=$'\t' read -r locale slug updated_at published_at robots canonical_url; do
    [ -n "$slug" ] || continue
    index=$((index + 1))

    local route_path="/$locale/blogs/$slug"
    local fetch_url="$base_url$route_path"
    local public_url="$SITE_URL$route_path"
    local expected_canonical
    expected_canonical="$(absolute_url "${canonical_url:-$route_path}")"
    local html_file="$TMP_DIR/verify-$index.html"
    local status

    status=""
    for _ in $(seq 1 12); do
      status="$(curl -sS -L --max-time 45 -o "$html_file" -w '%{http_code}' "$fetch_url" || true)"
      [ "$status" = "200" ] && break
      sleep 3
    done

    if [ "$status" != "200" ]; then
      log "verify failed: $public_url returned $status from $base_url"
      failures=$((failures + 1))
      continue
    fi

    if ! grep -Fq "<link rel=\"canonical\" href=\"$expected_canonical\"" "$html_file"; then
      log "verify failed: canonical mismatch for $public_url; expected $expected_canonical"
      failures=$((failures + 1))
    fi

    if [[ "$robots" != *noindex* ]]; then
      if grep -qi '<meta name="robots"[^>]*noindex' "$html_file"; then
        log "verify failed: unexpected noindex for $public_url"
        failures=$((failures + 1))
      elif ! grep -qi '<meta name="robots"[^>]*content="index,follow"' "$html_file"; then
        log "verify failed: missing index,follow robots tag for $public_url"
        failures=$((failures + 1))
      fi
    fi
  done <"$BLOGS_SORTED_TSV"

  if [ "$failures" -gt 0 ]; then
    return 1
  fi

  log "verified $index public blog route(s) against $base_url"
}

build_and_verify_image() {
  local cache_bust
  cache_bust="$(cat "$FINGERPRINT_FILE")"
  if [ "$FORCE" = "1" ]; then
    cache_bust="$cache_bust-force-$(date +%s%N)"
  fi

  log "building website image"
  (cd "$PROJECT_ROOT/deploy" && docker compose -f "$COMPOSE_FILE" build --build-arg "PRERENDER_CACHE_BUST=$cache_bust" website)

  local image_ref
  image_ref="$(cd "$PROJECT_ROOT/deploy" && docker compose -f "$COMPOSE_FILE" config --format json | jq -r '.services.website.image // empty')"
  if [ -z "$image_ref" ]; then
    image_ref="$WEBSITE_IMAGE_REF"
  fi
  BUILT_IMAGE_REF="$image_ref"
  if [ -z "$image_ref" ]; then
    log "could not resolve built website image reference"
    return 1
  fi

  local container_name="sangehassan-website-prerender-verify-$$"
  local container_id=""
  local verify_port=""

  cleanup_verify_container() {
    if [ -n "${container_id:-}" ]; then
      docker rm -f "$container_id" >/dev/null 2>&1 || true
    fi
  }

  container_id="$(docker run -d --rm --name "$container_name" -p 127.0.0.1::80 "$image_ref")"
  for _ in $(seq 1 20); do
    verify_port="$(docker port "$container_id" 80/tcp 2>/dev/null | sed -E 's/.*:([0-9]+)$/\1/' | tail -n 1)"
    [ -n "$verify_port" ] && break
    sleep 0.5
  done

  if [ -z "$verify_port" ]; then
    log "could not determine temporary website port"
    cleanup_verify_container
    return 1
  fi

  local verify_result=0
  verify_blog_routes "http://127.0.0.1:$verify_port" || verify_result=$?
  cleanup_verify_container
  return "$verify_result"
}

restore_previous_website() {
  local previous_image_id="$1"
  if [ -z "$previous_image_id" ] || [ -z "${BUILT_IMAGE_REF:-}" ]; then
    log "automatic rollback unavailable: previous or built image reference is missing"
    return 1
  fi

  log "restoring previous website image"
  docker tag "$previous_image_id" "$BUILT_IMAGE_REF"
  (cd "$PROJECT_ROOT/deploy" && docker compose -f "$COMPOSE_FILE" up -d --no-build --force-recreate website)
}

write_state() {
  mkdir -p "$STATE_DIR"
  {
    cat "$FINGERPRINT_FILE"
    printf '# refreshed_at=%s\n' "$(date -Is)"
    cat "$BLOGS_SORTED_TSV"
  } >"$STATE_FILE.tmp"
  mv "$STATE_FILE.tmp" "$STATE_FILE"
}

main() {
  need_cmd curl
  need_cmd docker
  need_cmd flock
  need_cmd jq
  need_cmd sha256sum

  mkdir -p "$(dirname "$LOCK_FILE")"
  exec 9>"$LOCK_FILE"
  if ! flock -n 9; then
    log "another blog prerender refresh is already running"
    exit 0
  fi

  TMP_DIR="$(mktemp -d)"
  BLOGS_TSV="$TMP_DIR/blogs.tsv"
  BLOGS_SORTED_TSV="$TMP_DIR/blogs.sorted.tsv"
  FINGERPRINT_FILE="$TMP_DIR/fingerprint"
  trap 'rm -rf "$TMP_DIR"' EXIT

  promote_due_scheduled_blogs
  fetch_public_blogs

  local current_fingerprint previous_fingerprint
  current_fingerprint="$(cat "$FINGERPRINT_FILE")"
  previous_fingerprint="$(sed -n '1p' "$STATE_FILE" 2>/dev/null || true)"

  if [ "$FORCE" != "1" ] && [ "$current_fingerprint" = "$previous_fingerprint" ]; then
    log "public blog fingerprint unchanged"
    exit 0
  fi

  log "public blog fingerprint changed; refreshing website prerender"
  local previous_image_id
  previous_image_id="$(docker inspect --format '{{.Image}}' sangehassan-website 2>/dev/null || true)"
  build_and_verify_image

  log "starting verified website image"
  if ! (cd "$PROJECT_ROOT/deploy" && docker compose -f "$COMPOSE_FILE" up -d --no-build website); then
    restore_previous_website "$previous_image_id" || true
    return 1
  fi

  if ! verify_blog_routes "$SITE_URL"; then
    restore_previous_website "$previous_image_id" || true
    return 1
  fi
  write_state
  log "blog prerender refresh completed"
}

main "$@"
