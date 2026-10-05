#!/usr/bin/env bash
# Minimal shell client for a running engine:  ./client.sh [port]
# The engine uses a self-signed loopback certificate, hence curl -k.
set -euo pipefail
PORT="${1:-20025}"
BASE="https://127.0.0.1:${PORT}/api/v1"
req() { curl -sk "$@"; echo; }

echo "== status";        req "$BASE/status"
echo "== capabilities";  req "$BASE/capabilities"
echo "== drm status";    req "$BASE/drm/status"

# Create a playback session for a catalog song, then stream it to a file.
ASSET="${ASSET:-}"
if [ -n "$ASSET" ]; then
  SESSION=$(curl -sk -X POST "$BASE/playback" -H 'Content-Type: application/json' \
    -d "{\"assetId\":\"$ASSET\",\"storefront\":\"us\",\"capabilities\":{\"lossless\":true}}")
  echo "== session: $SESSION"
  ID=$(printf '%s' "$SESSION" | sed -n 's/.*"sessionId":"\([^"]*\)".*/\1/p')
  curl -sk "$BASE/playback/$ID/audio" -o "track-$ASSET.m4a"
  curl -sk -X DELETE "$BASE/playback/$ID"
fi
