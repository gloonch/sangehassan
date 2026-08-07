#!/usr/bin/env bash
set -Eeuo pipefail

: "${CLOUDFLARE_API_TOKEN:?Set CLOUDFLARE_API_TOKEN with Zone Read, Bot Management Write, Cache Purge, and WAF Read permissions}"

CLOUDFLARE_ZONE_NAME="${CLOUDFLARE_ZONE_NAME:-sangehassan.com}"
CLOUDFLARE_API_BASE="${CLOUDFLARE_API_BASE:-https://api.cloudflare.com/client/v4}"
BACKUP_ROOT="${BACKUP_ROOT:-/opt/sangehassan/backups/cloudflare}"
BACKUP_DIR="$BACKUP_ROOT/$(date -u +%Y%m%dT%H%M%SZ)"
PROJECT_ROOT="${PROJECT_ROOT:-/opt/sangehassan}"
mkdir -p "$BACKUP_DIR"

api() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local args=(-fsS -X "$method" -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" -H "Content-Type: application/json")
  if [[ -n "$body" ]]; then args+=(-d "$body"); fi
  curl "${args[@]}" "$CLOUDFLARE_API_BASE$path"
}

if [[ -z "${CLOUDFLARE_ZONE_ID:-}" ]]; then
  zone_response="$(api GET "/zones?name=$CLOUDFLARE_ZONE_NAME&status=active")"
  printf '%s\n' "$zone_response" > "$BACKUP_DIR/zone-lookup.json"
  CLOUDFLARE_ZONE_ID="$(jq -er '.result[0].id' <<<"$zone_response")"
fi

bot_config="$(api GET "/zones/$CLOUDFLARE_ZONE_ID/bot_management")"
printf '%s\n' "$bot_config" > "$BACKUP_DIR/bot-management.before.json"
api GET "/zones/$CLOUDFLARE_ZONE_ID/rulesets" > "$BACKUP_DIR/rulesets.before.json"

desired_config='{
  "fight_mode": false,
  "ai_bots_protection": "disabled",
  "content_bots_protection": "disabled",
  "crawler_protection": "disabled",
  "is_robots_txt_managed": false,
  "cf_robots_variant": "off",
  "sbfm_verified_bots": "allow",
  "sbfm_likely_automated": "allow",
  "sbfm_definitely_automated": "allow",
  "sbfm_static_resource_protection": false
}'

update_response="$(api PUT "/zones/$CLOUDFLARE_ZONE_ID/bot_management" "$desired_config")"
printf '%s\n' "$update_response" > "$BACKUP_DIR/bot-management.update.json"
jq -e '.success == true' <<<"$update_response" >/dev/null

after_config="$(api GET "/zones/$CLOUDFLARE_ZONE_ID/bot_management")"
printf '%s\n' "$after_config" > "$BACKUP_DIR/bot-management.after.json"
jq -e '
  .success == true and
  .result.is_robots_txt_managed == false and
  (.result.ai_bots_protection == "disabled" or .result.ai_bots_protection == null) and
  (.result.crawler_protection == "disabled" or .result.crawler_protection == null) and
  (.result.fight_mode == false or .result.fight_mode == null)
' <<<"$after_config" >/dev/null

purge_payload="$(jq -cn --arg zone "$CLOUDFLARE_ZONE_NAME" '{files: [
  ("https://" + $zone + "/robots.txt"),
  ("https://" + $zone + "/sitemap.xml"),
  ("https://" + $zone + "/llms.txt"),
  ("https://" + $zone + "/fa/blogs"),
  ("https://" + $zone + "/en/blogs"),
  ("https://" + $zone + "/ar/blogs")
]}')"
purge_response="$(api POST "/zones/$CLOUDFLARE_ZONE_ID/purge_cache" "$purge_payload")"
printf '%s\n' "$purge_response" > "$BACKUP_DIR/cache-purge.json"
jq -e '.success == true' <<<"$purge_response" >/dev/null

if ! SINCE=5m "$PROJECT_ROOT/deploy/scripts/report-http-status-families.sh"; then
  restore_config="$(jq -c '.result | {
    fight_mode,
    ai_bots_protection,
    content_bots_protection,
    crawler_protection,
    is_robots_txt_managed,
    cf_robots_variant,
    sbfm_verified_bots,
    sbfm_likely_automated,
    sbfm_definitely_automated,
    sbfm_static_resource_protection
  } | with_entries(select(.value != null))' "$BACKUP_DIR/bot-management.before.json")"
  restore_response="$(api PUT "/zones/$CLOUDFLARE_ZONE_ID/bot_management" "$restore_config" || true)"
  printf '%s\n' "$restore_response" > "$BACKUP_DIR/bot-management.rollback.json"
  api POST "/zones/$CLOUDFLARE_ZONE_ID/purge_cache" "$purge_payload" > "$BACKUP_DIR/cache-purge.rollback.json" || true
  echo "Cloudflare verification failed; the previous Bot Management settings were restored from $BACKUP_DIR" >&2
  exit 1
fi

echo "Cloudflare AI crawler settings updated. Backup: $BACKUP_DIR"
echo "Review AI Crawl Control > Settings separately to confirm Pay Per Crawl is disabled; Cloudflare does not expose that control through the documented zone Bot Management API."
