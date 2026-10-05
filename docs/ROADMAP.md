# Roadmap and known issues

Work that is queued, in rough priority order.

## Known issues

1. **Embedded Android libraries break FairPlay** (see [drm.md](drm.md)). Decide whether to remove the embedding path or keep it as diagnostics only; shipping requires bundling `rootfs/system/lib64`.
2. **Widevine identity in git history.** The identity is no longer in the tree, but earlier commits contain it; rotate it or rewrite history before publishing.
3. **Bundled proprietary Android libraries.** `drm/rootfs/system/lib64` carries Apple's libraries (Git LFS). Keep the repository private, or ship an installer script instead; see [NOTICE.md](../NOTICE.md).

## Resolved

- **Music-video playback stopping after about a second.** Cause: the segment handler streamed past a fragment's end into the next `moof` when FFmpeg's output burst ahead of the fragment indexer, so the browser rejected the stray partial box (`CHUNK_DEMUXER_ERROR_APPEND_FAILED`). Fixed by bounding each fragment with the end of its own `mdat` (`MVLiveIndex.FragLimitByIndex`), capping the unknown-end frontier at the parsed bytes (`ParsedBytes`), and a 64 KiB safety margin; covered by `TestMVLiveIndex_FragLimitFromMdatBeforeNextMoof`.
- **Stripped open-source ICU libs failing to load** (`__register_atfork`). `drm/build-icu-ndk.sh` now retargets the import to `pthread_atfork` after linking.

- **Architecture tests were no-ops.** `sdk/archtest` still matched the old `engine/...` import paths, so no rule ever ran. It now uses the SDK module path, fails if package discovery returns too few packages, and was verified against a deliberate forbidden import.
- **OpenAPI coverage.** `api/openapi.json` describes all 69 routes, and `server/openapi_test.go` fails when a registered route is missing from the spec.
- **Vendored libhybris and bundled Android libraries.** The SDK builds and runs host-native FairPlay from a clean clone (libraries tracked with Git LFS).
- **Documentation accuracy.** `api.md` and `configuration.md` were checked against the handlers (for example `/library/sync` returns 410, VLC routes return 503 when unavailable).

- **DRM client TLS.** Server names are now verified against the certificate (previously only the chain was), and `MUSICKIT_TLS_PINS` enables optional SPKI SHA-256 pinning. No pins ship by default, so Apple certificate rotation cannot break playback.
- **Hard-coded `HiRes`.** It follows `BackendConfig.DisableHiRes` (`MUSICKIT_DISABLE_HIRES=1`). The account token carries no subscription tier, so hi-res is still assumed available otherwise.

- **Embedded Widevine identity.** Removed from `sdk/aacstream/cdm`; loaded from `MUSICKIT_WIDEVINE_DIR` instead.

- **Vendored wrapper licence.** Upstream is MIT; licence texts added under `drm/native/`.

- **MV seek latency (step 1).** The from-0 download is parked while a seek producer gets playable (`aacstream.HoldBackground`), because CDN bandwidth is shared across connections.

- **MV seek, fragment-level path.** Seeks fetch and decrypt only the fragment holding the target (about 1.4 s against 2.8 s for the whole segment on the real CDN), with automatic FFmpeg fallback; see [architecture.md](architecture.md#mv-seeking).

## Planned
- Validate fragment-level MV seeks against real Widevine-decrypted output in the player (the decrypt and the player's probe have only been exercised with synthetic and FFmpeg-built streams); consider promoting more of the seek flow into a small client library.

- Wire the vendored wrapper's login/2FA handlers to the DRM auth callback so a missing Android session can be created from the engine.
- Surface lease-recovery state (`Running`, `Scheduled`, `Refreshing`, `Failed`) in `/drm/status` and the SSE `drm` event.
- Progressive MV (itun) decryption through the in-process library.
- Split `server` handlers into focused packages behind small interfaces so individual areas (library, export, catalog) can be embedded without the rest.
- Provide a real recommendations client example that renders artwork and deep links.
- CI: build, `go test ./...` in all modules, `sdk/archtest`, `gofmt` check.
