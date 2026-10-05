# drm

`libdrm_client.so` — the in-process DRM library the Go `sdk/drm` package links through cgo.

```sh
make libdrm_client.so      # embeds the vendored libhybris-core.so
make android-stubs         # generates libandroid.so / libOpenSLES.so in rootfs/system/lib64
```

Files: `drm_client.*` (clean-room client), `drm_hybris.*` (bridge to the Android FairPlay libraries), `embedded_loader.*` (loads the embedded libhybris-core.so), `native/` (vendored wrapper in library mode, see [../NOTICE.md](../NOTICE.md)), `build-icu-ndk.sh` (optional ICU build), `libhybris-core.so` + `hybris-linker/q.so` + `vendor/libhybris/` (vendored libhybris with provenance, patch, licences and rebuild script).

Runtime requirements (Android libraries, `hybris-linker/q.so`), the key and lease flow, and failure modes are in [../docs/drm.md](../docs/drm.md).
