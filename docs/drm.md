# DRM client

`drm/` builds `libdrm_client.so`, the in-process library the engine links through cgo (build tag `native_backend`). Without that tag `NewNativeBackend` is a stub: DRM is disabled and the engine falls back to AAC.

## Components

| File | Role |
|---|---|
| `drm_client.c/.h`, `drm_types.h` | Clean-room client: HTTPS (with HTTP/2 option), cookie jar, login and 2FA, token handling, license pool, key-context cache, itun decryption entry points, state and auth callbacks |
| `drm_hybris.c/.h` | Bridge to the Android FairPlay libraries: environment setup, library loading, and delegation of init, key derivation and decrypt to the native wrapper |
| `embedded_loader.c/.h` | Optional in-memory loader for the Android libraries (see below) |
| `native/` | Vendored host-native wrapper in library mode: request-context setup, lease management and recovery, key contexts, exception barrier, shims onto the loader |
| `build-icu-ndk.sh` | Builds ICU for Android x86_64 (optional; Apple's own ICU libs also work) |

The "clean-room" part is `drm_client.*`: written from observed protocol behaviour. The FairPlay key exchange itself cannot be reimplemented, so it runs Apple's own Android libraries on Linux through libhybris; `native/` is the third-party reference wrapper that drives them (see [NOTICE.md](../NOTICE.md)).

## Session and key flow

1. `drm_init` loads cached credentials from the DRM base directory (`<base>/mpl_db`, `MUSIC_TOKEN`, `STOREFRONT_ID`), starts the HTTPS stack, and reports `STARTING`, `LOGIN`, `WAITING_2FA`, `INITIALIZING_FAIRPLAY`, `RUNNING` or `FAILED` through the state callback.
2. The Go backend then initialises hybris and calls `hybris_fairplay_init`, which runs the wrapper's `drm_lib_init`: configure the device identity and request context from the 9-field device string, acquire the playback lease, cache account tokens.
3. For each `skd://` key URI the engine calls `drm_open_key_context(adamId, uri)`. The wrapper runs `getPersistentKey` (the exchange with Apple's key server) and derives a decrypt handle. Prefetch key `skd://itunes.apple.com/P000000000/s1/e1` (adam `0`) is shared and cached.
4. `drm_decrypt_sample*` decrypts whole 16-byte blocks in place with selector 5.
5. There is **no zero-key fallback**: if a real key context cannot be produced, opening fails and the stream errors instead of emitting garbage audio.

## Lease management and recovery

The lease manager is created with real callbacks. A lease end or playback error schedules recovery on a worker thread: events are coalesced, retried with exponential backoff (1, 2, 5, 10, then 30 s repeating), and key requests are refused while recovery is active. Recovery resets all FairPlay contexts and rebuilds the prefetch context. Every refresh increments an epoch, and cached key contexts from an older epoch are discarded rather than reused. The states are `Running`, `Scheduled`, `Refreshing`, `Failed`.

## Runtime layout

libhybris ships in this repository (`drm/libhybris-core.so`, `drm/hybris-linker/q.so`; see `drm/vendor/libhybris/README.md`). The Android and Apple Music native libraries are included in `drm/rootfs/system/lib64` (proprietary; see [NOTICE.md](../NOTICE.md)); `scripts/install-android-libs.sh` replaces them with another set. The runtime layout is:

```
<drm dir>/hybris-linker/q.so              hybris linker plugin (vendored)
<drm dir>/rootfs/system/lib64/*.so        Android system + Apple Music libs (included; must include libc.so)
<drm dir>/files/                          session data (created at runtime): mpl_db/, MUSIC_TOKEN, …
```

The engine searches for `lib64` in `$MUSICKIT_DRM_DIR/rootfs/system/lib64`, `<drm dir>/rootfs/system/lib64`, `<parent>/rootfs/system/lib64`, then `~/.config/musickit-sdk-linux/drm/rootfs/system/lib64`, taking the first that contains `libc.so`. The linker directory is the first of those roots that contains `hybris-linker/q.so`. Hybris is configured automatically (`HYBRIS_LINKER_DIR`, `HYBRIS_LD_LIBRARY_PATH`, `HYBRIS_ANDROID_LIB64`).

By default the Android libraries load **from disk**. Loading them from in-memory copies (`MUSICKIT_EMBED_LIBS=1`) breaks FairPlay: libraries that locate files relative to their own path make the lease request fail. The embedded blobs remain buildable but are opt-in.

## Failure modes

| Log line | Meaning |
|---|---|
| `hybris backend: not available (linker=… lib64=…)` | No usable `q.so` or `lib64` was found |
| `hybris key context unavailable … refusing zero-key fallback` | Key exchange failed; check recovery logs and the session |
| `[guard] … exception` | A C++ exception from the Android libraries was caught and converted to an error |
| `key context refused: lease recovery in progress` | Transient; retry after the backoff |
| `cannot locate symbol "__register_atfork"` | An NDK-built ICU lib imports a symbol the Android libc lacks; use Apple's ICU libs or patch the import to `pthread_atfork` (same length, NUL padded) |

## Testing

`go test ./sdk/drm` covers the manager, state machine and the CBCS protocol server with a fake backend. The C library has no unit tests; verify with a real playback and the log lines above.
