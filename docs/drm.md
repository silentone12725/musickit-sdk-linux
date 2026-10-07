# DRM client

`drm/` builds `libdrm_client.so` (`libdrm_client.dll` on Windows), the in-process library the engine links through cgo (build tag `widevine_backend`). Without that tag `WidevineBackend` is a stub: DRM is disabled and the engine falls back to AAC.

## Components

| File | Role |
|---|---|
| `drm_linux.c` | The flat `drm_*` API the Go backend calls: session directory, account lookup, key contexts, AES-128-CBC sample decryption (OpenSSL `EVP`). The only file built into the Windows DLL |
| `drm_client.*`, `drm_types.h` | Clean-room client types and helpers shared with the Linux build |
| `fairplay*.c`, `fp_*.c` | Clean-room FairPlay building blocks (SPC/CKC parsing, HTTP exchange); built on Linux, not on the playback path |

## Session and key flow

1. `drm_init` records the session directory (`<base>/mpl_db`, `MUSIC_TOKEN`, `STOREFRONT_ID`) and reports `RUNNING` through the state callback. Credentials passed to it are ignored.
2. For each `skd://` key URI the engine acquires the content key in Go: `aacstream.AcquireKey` builds a Widevine PSSH, runs the software CDM, posts the challenge to Apple's `acquireWebPlaybackLicense` endpoint with the developer token and the user's music token, and unwraps the returned licence to a 16-byte AES key.
3. The backend opens a key context (`drm_open_key_context`), injects the key (`drm_set_key_context_key`) and decrypts samples in place with `drm_decrypt_sample_at` (whole 16-byte blocks, CBC, zero IV for CBCS ALAC).
4. There is **no zero-key fallback**: if a real key cannot be obtained, opening fails and the stream errors instead of emitting garbage audio.

## Layout

```
<drm dir>/libdrm_client.so|.dll    next to the engine binary ($ORIGIN rpath on Linux, exe directory on Windows)
<drm dir>/files/                   session data (created at runtime): mpl_db/, MUSIC_TOKEN, STOREFRONT_ID, …
```

On Windows the DLL needs `libcrypto-3-x64.dll` and `libwinpthread-1.dll` from the MSYS2 UCRT64 `bin/` directory beside it.

## Failure modes

| Log line | Meaning |
|---|---|
| `DRM backend name=widevine` absent, `lossless=false` | The engine was built without `-tags widevine_backend` |
| `key context refused …` | Key acquisition failed; check the session (`/api/v1/drm/status`) |
| `session: empty` in capabilities | No `MUSIC_TOKEN` / `STOREFRONT_ID` in the session directory yet |

## Testing

`go test ./sdk/drm` covers the manager, state machine and the CBCS protocol server with a fake backend. The C library has no unit tests; verify with a real playback.
