# Building

## Prerequisites

| Need | For |
|---|---|
| Go 1.25+ | all Go modules |
| gcc and make | `drm/` |
| OpenSSL development files (and libcurl on Linux) | `drm/` |
| libvlc development files | `sdk/vlc` (in-process playback) |
| ffmpeg, MP4Box (optional) | export, video remux |

No proprietary or Android libraries are involved: content keys come from Widevine in Go, and the C library only needs OpenSSL.

## Steps (Linux)

```sh
# 1. DRM client library
make -C drm libdrm_client.so

# 2. Engine binary with real DRM (CGO + widevine_backend tag)
scripts/build-engine.sh dist            # → dist/musickit-engine + dist/libdrm_client.so

# 3. Run
cd <dir containing drm/>
./dist/musickit-engine --api 20025
```

The engine binary finds `libdrm_client.so` next to itself (`$ORIGIN` rpath). `scripts/build-engine.sh` links through a temporary spaceless directory because cgo cannot take `-L` paths containing spaces.

## Steps (Windows)

Install [MSYS2](https://www.msys2.org), then in a UCRT64 shell:

```sh
pacman -S mingw-w64-ucrt-x86_64-{gcc,make,openssl}
```

and put `C:\msys64\ucrt64\bin` on `PATH`. Build the DLL and the engine (VLC's headers come from the VLC source tree, its DLLs from a VLC install):

```sh
mingw32-make -C drm libdrm_client.dll
cd server
set CGO_ENABLED=1
set CGO_CFLAGS=-I<vlc-src>/include
set CGO_LDFLAGS=-L<VLC install dir> -L<repo>/drm
go build -tags widevine_backend -o ..\dist\musickit-engine.exe .\cmd\musickit-engine
```

Run it with the VLC install directory, `drm\` and `C:\msys64\ucrt64\bin` on `PATH` (or copy the DLLs beside the executable). The Cavern Atmos bridge is Linux-only; on Windows binaural rendering uses the FFmpeg path.

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

## Cleaning

```sh
scripts/clean.sh --dry-run   # list what would go
scripts/clean.sh             # objects, test binaries, stray go build output
scripts/clean.sh --all       # also the built library and dist/
make clean                   # the same intermediates, through make
make distclean               # everything built
make release                 # build the real engine, then drop the intermediates
```

`drm/files` (an Apple session) is never removed.
