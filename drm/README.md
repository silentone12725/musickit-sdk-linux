# drm

`libdrm_client.so` — the in-process DRM library the Go `sdk/drm` package links through cgo.

```sh
export HYBRIS_CORE=/path/to/libhybris-core.so
make libdrm_client.so
```

Files: `drm_client.*` (clean-room client), `drm_hybris.*` (bridge to the Android FairPlay libraries), `embedded_loader.*` (optional in-memory loader), `native/` (vendored wrapper in library mode, see [../NOTICE.md](../NOTICE.md)), `build-icu-ndk.sh` (optional ICU build).

Runtime requirements (Android libraries, `hybris-linker/q.so`), the key and lease flow, and failure modes are in [../docs/drm.md](../docs/drm.md).
