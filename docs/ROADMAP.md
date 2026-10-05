# Roadmap and known issues

Work that is queued, in rough priority order.

## Known issues

1. **Embedded Android libraries break FairPlay** (see [drm.md](drm.md)). Decide whether to remove the embedding path or keep it as diagnostics only; shipping requires bundling `rootfs/system/lib64`.
2. **Embedded Widevine identity** in `sdk/aacstream/cdm/consts.go` (see [security.md](security.md)). Move to configuration before distributing.
3. **No certificate pinning** for the DRM client's HTTPS requests (TODO in `drm/drm_client.c`).
4. **`HiRes: true` is hardcoded** in `sdk/drm/manager.go`; it should come from configuration.
5. **Bundled proprietary Android libraries.** `drm/rootfs/system/lib64` carries Apple's libraries (Git LFS). Keep the repository private, or ship an installer script instead; see [NOTICE.md](../NOTICE.md).
6. **Vendored wrapper licence** has not been verified.

## Resolved

- **Music-video playback stopping after about a second.** Cause: the segment handler streamed past a fragment's end into the next `moof` when FFmpeg's output burst ahead of the fragment indexer, so the browser rejected the stray partial box (`CHUNK_DEMUXER_ERROR_APPEND_FAILED`). Fixed by bounding each fragment with the end of its own `mdat` (`MVLiveIndex.FragLimitByIndex`), capping the unknown-end frontier at the parsed bytes (`ParsedBytes`), and a 64 KiB safety margin; covered by `TestMVLiveIndex_FragLimitFromMdatBeforeNextMoof`.
- **Stripped open-source ICU libs failing to load** (`__register_atfork`). `drm/build-icu-ndk.sh` now retargets the import to `pthread_atfork` after linking.

- **Architecture tests were no-ops.** `sdk/archtest` still matched the old `engine/...` import paths, so no rule ever ran. It now uses the SDK module path, fails if package discovery returns too few packages, and was verified against a deliberate forbidden import.
- **OpenAPI coverage.** `api/openapi.json` describes all 69 routes, and `server/openapi_test.go` fails when a registered route is missing from the spec.
- **Vendored libhybris and bundled Android libraries.** The SDK builds and runs host-native FairPlay from a clean clone (libraries tracked with Git LFS).
- **Documentation accuracy.** `api.md` and `configuration.md` were checked against the handlers (for example `/library/sync` returns 410, VLC routes return 503 when unavailable).

## Planned

- Wire the vendored wrapper's login/2FA handlers to the DRM auth callback so a missing Android session can be created from the engine.
- Surface lease-recovery state (`Running`, `Scheduled`, `Refreshing`, `Failed`) in `/drm/status` and the SSE `drm` event.
- Progressive MV (itun) decryption through the in-process library.
- Split `server` handlers into focused packages behind small interfaces so individual areas (library, export, catalog) can be embedded without the rest.
- Provide a real recommendations client example that renders artwork and deep links.
- CI: build, `go test ./...` in all modules, `sdk/archtest`, `gofmt` check.
