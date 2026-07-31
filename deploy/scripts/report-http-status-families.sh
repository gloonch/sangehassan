#!/usr/bin/env bash
set -Eeuo pipefail

NGINX_CONTAINER="${NGINX_CONTAINER:-sangehassan-nginx}"
SINCE="${SINCE:-24h}"

docker logs --since "$SINCE" "$NGINX_CONTAINER" 2>&1 | awk '
  {
    split($0, quoted, "\"")
    split(quoted[2], request, " ")
    split(quoted[3], response, " ")
    path=request[2]
    code=response[2]
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
' | sort
