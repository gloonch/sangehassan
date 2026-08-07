#!/usr/bin/env bash
set -Eeuo pipefail

NGINX_CONTAINER="${NGINX_CONTAINER:-sangehassan-nginx}"
SINCE="${SINCE:-24h}"
SITE_URL="${SITE_URL:-https://sangehassan.com}"
AI_CHECK_ARTICLE_PATH="${AI_CHECK_ARTICLE_PATH:-/fa/blogs/everything-about-travertine-stone}"
AI_CRAWLERS=("ChatGPT-User" "OAI-SearchBot" "GPTBot" "ClaudeBot" "PerplexityBot")
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

docker logs --since "$SINCE" "$NGINX_CONTAINER" 2>&1 | awk '
  {
    split($0, quoted, "\"")
    split(quoted[2], request, " ")
    response=quoted[3]
    sub(/^[[:space:]]+/, "", response)
    split(response, response_parts, /[[:space:]]+/)
    path=request[2]
    code=response_parts[1]
    if (path == "" || code !~ /^[1-5][0-9][0-9]$/) next
    sub(/\?.*$/, "", path)
    split(path, parts, "/")
    family="/"
    if (parts[2] != "") family="/" parts[2]
    if (parts[3] != "") family=family "/" parts[3]
    counts[substr(code, 1, 1) "xx\t" family]++
    if (code ~ /^5/) errors[code "\t" path]++
  }
  END {
    print "status_family\tpath_family\tcount"
    for (key in counts) print key "\t" counts[key]
    print ""
    print "5xx_status\tpath\tcount"
    for (key in errors) print key "\t" errors[key]
  }
'

echo
echo "ai_crawler\tstatus\tpath_family\tcount"
docker logs --since "$SINCE" "$NGINX_CONTAINER" 2>&1 | awk '
  {
    split($0, quoted, "\"")
    split(quoted[2], request, " ")
    response=quoted[3]
    sub(/^[[:space:]]+/, "", response)
    split(response, response_parts, /[[:space:]]+/)
    path=request[2]
    code=response_parts[1]
    ua=quoted[6]
    crawler=""
    if (ua ~ /ChatGPT-User/) crawler="ChatGPT-User"
    else if (ua ~ /OAI-SearchBot/) crawler="OAI-SearchBot"
    else if (ua ~ /GPTBot/) crawler="GPTBot"
    else if (ua ~ /ClaudeBot/) crawler="ClaudeBot"
    else if (ua ~ /PerplexityBot/) crawler="PerplexityBot"
    if (crawler == "" || path == "" || code !~ /^[1-5][0-9][0-9]$/) next
    sub(/\?.*$/, "", path)
    split(path, parts, "/")
    family="/"
    if (parts[2] != "") family="/" parts[2]
    if (parts[3] != "") family=family "/" parts[3]
    counts[crawler "\t" code "\t" family]++
  }
  END {
    for (key in counts) print key "\t" counts[key]
  }
'

robots_file="$TMP_DIR/robots.txt"
robots_headers="$TMP_DIR/robots.headers"
robots_status="$(curl -sS -L -D "$robots_headers" -o "$robots_file" -w '%{http_code}' "$SITE_URL/robots.txt")"
if [[ "$robots_status" != "200" ]] || grep -qi '^cf-mitigated:[[:space:]]*challenge' "$robots_headers"; then
  echo "AI synthetic check failed: robots.txt returned $robots_status or a Cloudflare challenge" >&2
  exit 1
fi

python3 - "$robots_file" "$SITE_URL$AI_CHECK_ARTICLE_PATH" "$SITE_URL/api/admin/blogs" "${AI_CRAWLERS[@]}" <<'PY'
import sys
import urllib.robotparser

robots_path, public_url, private_url, *agents = sys.argv[1:]
parser = urllib.robotparser.RobotFileParser()
with open(robots_path, encoding="utf-8") as handle:
    parser.parse(handle.read().splitlines())

errors = []
for agent in agents:
    if not parser.can_fetch(agent, public_url):
        errors.append(f"{agent} is disallowed from public articles")
    if parser.can_fetch(agent, private_url):
        errors.append(f"{agent} is allowed on private API paths")
if errors:
    raise SystemExit("AI robots verification failed:\n" + "\n".join(errors))
PY

for crawler in "${AI_CRAWLERS[@]}"; do
  for path in "/fa/blogs" "$AI_CHECK_ARTICLE_PATH"; do
    safe_name="${crawler//[^A-Za-z0-9]/_}-${path//\//_}"
    headers="$TMP_DIR/$safe_name.headers"
    body="$TMP_DIR/$safe_name.html"
    status="$(curl -sS -L -A "$crawler" -D "$headers" -o "$body" -w '%{http_code}' "$SITE_URL$path")"
    if [[ "$status" != "200" ]]; then
      echo "AI synthetic check failed: $crawler received $status for $path" >&2
      exit 1
    fi
    if grep -qi '^cf-mitigated:[[:space:]]*challenge' "$headers"; then
      echo "AI synthetic check failed: Cloudflare challenged $crawler on $path" >&2
      exit 1
    fi
    if ! grep -qi '^content-type:.*text/html' "$headers" || ! grep -Eqi '<h1([[:space:]>])' "$body"; then
      echo "AI synthetic check failed: incomplete HTML for $crawler on $path" >&2
      exit 1
    fi
  done
done

echo "AI synthetic checks passed for ${#AI_CRAWLERS[@]} crawler user agents."
