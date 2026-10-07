# drm

`libdrm_client.so` (`libdrm_client.dll` on Windows) — the in-process DRM library the Go `sdk/drm` package links through cgo.

```sh
make libdrm_client.so       # Linux: needs OpenSSL and libcurl development files
mingw32-make libdrm_client.dll   # Windows (MSYS2 UCRT64): builds drm_linux.c against libcrypto only
```

Files: `drm_linux.c` (the flat `drm_*` API the engine calls: key contexts and AES-128-CBC decryption over OpenSSL), `drm_client.*` and `drm_types.h` (clean-room client), `fairplay*.c`, `fp_*.c` (clean-room FairPlay building blocks, not on the playback path).

Content keys are acquired in Go through Widevine (`sdk/aacstream`); this library only holds key contexts and decrypts. See [../docs/drm.md](../docs/drm.md).
