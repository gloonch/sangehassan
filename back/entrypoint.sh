#!/bin/sh
set -eu

# If a read-only seed directory is mounted (prod compose), copy it into the
# persistent images volume once. This prevents missing built-in images like
# /images/templates/*.png after switching to a named volume.
if [ -d /seed/images ]; then
  mkdir -p /app/storage/images
  # If the volume is fully empty, copy everything.
  if ! find /app/storage/images -type f -maxdepth 3 2>/dev/null | grep -q .; then
    cp -R /seed/images/. /app/storage/images/
  else
    # Otherwise, ensure key subtrees exist (templates, products) without overwriting uploads.
    if [ -d /seed/images/templates ] && { [ ! -d /app/storage/images/templates ] || [ -z "$(ls -A /app/storage/images/templates 2>/dev/null || true)" ]; }; then
      cp -R /seed/images/templates /app/storage/images/
    fi
    if [ -d /seed/images/products ] && { [ ! -d /app/storage/images/products ] || [ -z "$(ls -A /app/storage/images/products 2>/dev/null || true)" ]; }; then
      cp -R /seed/images/products /app/storage/images/
    fi
    if [ -d /seed/images/watermark ]; then
      mkdir -p /app/storage/images/watermark
      cp -R /seed/images/watermark/. /app/storage/images/watermark/
    fi
    # Provide a few generic placeholder images under /images/content/.
    mkdir -p /app/storage/images/content
    for f in /seed/images/granite.jpg /seed/images/marble.jpg /seed/images/travertine.jpg; do
      if [ -f "$f" ]; then
        base="$(basename "$f")"
        if [ ! -f "/app/storage/images/content/$base" ]; then
          cp "$f" "/app/storage/images/content/$base"
        fi
      fi
    done
  fi
fi

exec /app/server
