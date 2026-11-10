#!/bin/sh
# Download Apple's Android FairPlay libraries needed by a backend that runs them
# (for example the app's hybris backend). They cannot be redistributed, so nothing
# is bundled: they are fetched from the sources pinned in drm/fairplay-libs.lock and
# each file is checked against its SHA-256 before use. The SDK's built-in Widevine
# backend does not need them.
#
# Usage: fetch-fairplay-libs.sh [--dest DIR] [--lock FILE] [--stubs DIR] [--quiet]
#
#   --dest   where the libraries go (default:
#            ${XDG_CONFIG_HOME:-~/.config}/musickit-sdk-linux/drm/rootfs/system/lib64)
#   --lock   the pinned source/checksum list (default: ../drm/fairplay-libs.lock)
#   --stubs  optional directory holding libandroid/libOpenSLES/libdl stubs to copy
#            next to the libraries (skipped when not given)
#
# Exit status: 0 when every library is in place and verified, 1 otherwise.
# Safe to re-run: files that already match their checksum are not downloaded again.
set -u

here=$(cd "$(dirname "$0")" && pwd)
DEST="${XDG_CONFIG_HOME:-$HOME/.config}/musickit-sdk-linux/drm/rootfs/system/lib64"
LOCK="$here/../drm/fairplay-libs.lock"
STUBS=""
QUIET=0
MAX_BYTES=67108864 # 64 MiB; the largest library is ~14 MiB

while [ $# -gt 0 ]; do
    case "$1" in
        --dest)  DEST="$2"; shift 2 ;;
        --lock)  LOCK="$2"; shift 2 ;;
        --stubs) STUBS="$2"; shift 2 ;;
        --quiet) QUIET=1; shift ;;
        -h|--help) sed -n '2,15p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
        *) echo "fetch-fairplay-libs: unknown option $1" >&2; exit 2 ;;
    esac
done

say() { [ "$QUIET" = 1 ] || echo "$@"; }
err() { echo "fetch-fairplay-libs: $*" >&2; }

[ -r "$LOCK" ] || { err "lock file $LOCK not found"; exit 1; }

if command -v sha256sum >/dev/null 2>&1; then
    sha256() { sha256sum "$1" | cut -d' ' -f1; }
elif command -v shasum >/dev/null 2>&1; then
    sha256() { shasum -a 256 "$1" | cut -d' ' -f1; }
else
    err "need sha256sum or shasum to verify downloads"; exit 1
fi

if command -v curl >/dev/null 2>&1; then
    fetch() { curl --proto '=https' --tlsv1.2 -fsSL --retry 3 --connect-timeout 20 --max-time 600 \
                   --max-filesize "$MAX_BYTES" -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
    fetch() { wget -q --https-only --tries=3 --timeout=60 -O "$2" "$1" &&
              [ "$(wc -c < "$2")" -le "$MAX_BYTES" ]; }
else
    err "need curl or wget to download the libraries"; exit 1
fi

mkdir -p "$DEST" || { err "cannot create $DEST"; exit 1; }

sources=$(sed -n 's/^source[[:space:]]\{1,\}//p' "$LOCK")
[ -n "$sources" ] || { err "no sources in $LOCK"; exit 1; }

total=0; fetched=0; failed=""
# Entries are "<sha256>  <filename>"; comments and source lines are skipped.
entries=$(grep -E '^[0-9a-f]{64}[[:space:]]+[A-Za-z0-9._+-]+\.so$' "$LOCK")
[ -n "$entries" ] || { err "no library entries in $LOCK"; exit 1; }

old_ifs=$IFS
IFS='
'
for line in $entries; do
    IFS=$old_ifs
    want=${line%% *}
    name=${line##* }
    total=$((total + 1))
    target="$DEST/$name"
    if [ -f "$target" ] && [ "$(sha256 "$target")" = "$want" ]; then
        continue
    fi
    part="$DEST/.$name.part"
    ok=0
    for src in $sources; do
        rm -f "$part"
        if fetch "$src/$name" "$part" && [ "$(sha256 "$part")" = "$want" ]; then
            chmod 644 "$part" && mv -f "$part" "$target" && ok=1 && break
        fi
        err "$name from $src failed or did not match its checksum"
    done
    rm -f "$part"
    if [ "$ok" = 1 ]; then
        fetched=$((fetched + 1)); say "  $name"
    else
        failed="$failed $name"
    fi
    IFS='
'
done
IFS=$old_ifs

for stub in libandroid.so libOpenSLES.so libdl.so; do
    [ -n "$STUBS" ] || break
    if [ -f "$STUBS/$stub" ]; then
        cp -f "$STUBS/$stub" "$DEST/$stub" && chmod 644 "$DEST/$stub"
    elif [ ! -f "$DEST/$stub" ]; then
        failed="$failed $stub"
        err "stub $stub not found in $STUBS"
    fi
done

if [ -n "$failed" ]; then
    err "missing:$failed — the backend that needs these libraries stays off until this succeeds (re-run this script)"
    exit 1
fi
say "FairPlay libraries ready in $DEST ($fetched of $total downloaded, the rest already current)"
exit 0
