# Building

## Prerequisites

| Need | For |
|---|---|
| Go 1.25+ | all Go modules |
| gcc, g++, make, GNU as, xxd | `drm/` |
| OpenSSL and libcurl development files | `drm/` |
| libvlc development files | `sdk/vlc` (in-process playback) |
| Android rootfs libraries (`libc.so`, Apple Music libs, …) | FairPlay (runtime and `drm/` build); **included** in `drm/rootfs/system/lib64`; replace with `scripts/install-android-libs.sh` |
| ffmpeg, MP4Box (optional) | export, video remux |

libhybris is **included**: `drm/libhybris-core.so` and `drm/hybris-linker/q.so` are prebuilt, with source provenance, patch, licences and a rebuild script in `drm/vendor/libhybris/`. The Android and Apple Music native libraries are also **included** (`drm/rootfs/system/lib64`, 99 files, ~116 MB). They are proprietary third-party binaries; see [NOTICE.md](../NOTICE.md). To use a different set, run `scripts/install-android-libs.sh` (below).

## Steps

```sh
# 0. (optional) replace the bundled Android / Apple Music libraries
scripts/install-android-libs.sh /path/to/lib64          # or a rootfs dir, or a tarball

# 1. DRM client library (uses the vendored libhybris-core.so)
make -C drm libdrm_client.so

# 2. Engine binary with real DRM (CGO + native_backend tag)
scripts/build-engine.sh dist            # → dist/musickit-engine + dist/libdrm_client.so

# 3. Run
cd <dir containing drm/>
./dist/musickit-engine --api 20025
```

The engine binary finds `libdrm_client.so` next to itself (`$ORIGIN` rpath). Override `HYBRIS_CORE=<relative path>` to embed your own libhybris build. `scripts/build-engine.sh` links through a temporary spaceless directory because cgo cannot take `-L` paths containing spaces.

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

## Cleaning

A build leaves intermediates behind; the DRM library alone generates about 400 MB of embedded-blob sources.

```sh
scripts/clean.sh --dry-run   # list what would go
scripts/clean.sh             # objects, embedded blobs, probe/test binaries, stray go build output
scripts/clean.sh --all       # also drm/libdrm_client.so and dist/
make clean                   # the same intermediates, through make
make distclean               # everything built
make release                 # build the real engine, then drop the intermediates
make -C drm release          # build only the library, then drop what it was built from
```

The next `make` regenerates the blobs, which takes a minute or two. `drm/files` (an Apple session), `drm/rootfs` and the vendored libhybris are never removed.
