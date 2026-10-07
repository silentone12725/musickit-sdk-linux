#!/usr/bin/env bash
# clean.sh — remove generated build output.
#
#   scripts/clean.sh             intermediates: compiler objects, test binaries, Go test binaries,
#                                binaries `go build` left in the source tree
#   scripts/clean.sh --all       also the built library (drm/libdrm_client.so) and dist/
#   scripts/clean.sh --dry-run   list what would be removed, remove nothing (combine with --all)
#
# Never touched: sources, drm/files (an Apple session).
# `make clean` / `make distclean` do the same through make.

set -euo pipefail

REPO="$(cd "$(dirname "$0")/.." && pwd)"
DRY=0; ALL=0
for arg in "$@"; do
    case "$arg" in
        -n|--dry-run) DRY=1 ;;
        --all)        ALL=1 ;;
        -h|--help)    sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
        *) echo "unknown option: $arg (see --help)" >&2; exit 2 ;;
    esac
done

targets=(
    examples/go-recommendations/go-recommendations server/musickit-engine sdk/musickit-engine
)
while IFS= read -r -d '' f; do targets+=("${f#"$REPO"/}"); done < <(
    find "$REPO/drm" "$REPO/sdk" "$REPO/server" "$REPO/examples" \( -name '*.o' -o -name '*.test' \) -type f \
        -print0)
for f in "$REPO"/drm/tests/test_*; do
    [ -f "$f" ] && [ -x "$f" ] && [[ "$f" != *.c && "$f" != *.py && "$f" != *.sh ]] && targets+=("${f#"$REPO"/}")
done
if [ "$ALL" = 1 ]; then
    targets+=(dist drm/libdrm_client.so drm/libdrm_client.dll drm/libdrm_client.dll.a)
fi

total=0
for rel in "${targets[@]}"; do
    path="$REPO/$rel"
    [ -e "$path" ] || continue
    kb=$(du -sk "$path" | cut -f1)
    total=$((total + kb))
    printf '%s %8s  %s\n' "$([ "$DRY" = 1 ] && echo would-remove || echo removed)" "$((kb / 1024))M" "$rel"
    [ "$DRY" = 1 ] || rm -rf "$path"
done
echo "$([ "$DRY" = 1 ] && echo 'would free' || echo 'freed') $((total / 1024)) MB"
