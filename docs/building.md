# Building

## Prerequisites

| Need | For |
|---|---|
| Go 1.25+ | all Go modules |
| gcc, g++, make, GNU as, xxd | `drm/` |
| OpenSSL and libcurl development files | `drm/` |
| libvlc development files | `sdk/vlc` (in-process playback) |
| libhybris-core build and its linker plugin `q.so` | FairPlay (runtime and `drm/` build) |
| Android rootfs libraries (`libc.so`, Apple Music libs, …) | FairPlay (runtime; not distributed) |
| ffmpeg, MP4Box (optional) | export, video remux |

Sourcing libhybris and the Android libraries is your responsibility; this repository does not ship them.

## Steps

```sh
# 1. DRM client library
export HYBRIS_CORE=/path/to/libhybris-core.so
make -C drm libdrm_client.so

# 2. Engine binary with real DRM (CGO + native_backend tag)
scripts/build-engine.sh dist            # → dist/musickit-engine + dist/libdrm_client.so

# 3. Run
cd <dir containing drm/>
./dist/musickit-engine --api 20025
```

The engine binary finds `libdrm_client.so` next to itself (`$ORIGIN` rpath). `scripts/build-engine.sh` links through a temporary spaceless directory because cgo cannot take `-L` paths containing spaces.

### Without DRM

```sh
cd server && go build -o ../dist/musickit-engine ./cmd/musickit-engine
```

This builds without cgo DRM: lossless and DRM-protected playback are disabled, but catalog, lyrics, metadata and the personalised feeds work.

### Tests and checks

```sh
make test     # sdk, server and examples
make vet
make fmt      # lists unformatted files
```

`sdk/archtest` verifies the import boundaries; run it after moving packages (`cd sdk && go test ./archtest`).

### Optional: ICU

`drm/build-icu-ndk.sh` builds ICU 68.2 for Android x86_64 in Docker (NDK r23b) and strips unused locales with `icupkg`. Apple's original ICU libraries are smaller and work out of the box.
