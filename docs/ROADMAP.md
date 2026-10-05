# Roadmap and known issues

Work that is queued, in rough priority order.

## Known issues

1. **Embedded Android libraries break FairPlay** (see [drm.md](drm.md)). Decide whether to remove the embedding path or keep it as diagnostics only; shipping requires bundling `rootfs/system/lib64`.
2. **Embedded Widevine identity** in `sdk/aacstream/cdm/consts.go` (see [security.md](security.md)).

## Resolved

- **Music-video playback stopping after about a second.** Cause: the segment handler streamed past a fragment's end into the next `moof` when FFmpeg's output burst ahead of the fragment indexer, so the browser rejected the stray partial box (`CHUNK_DEMUXER_ERROR_APPEND_FAILED`). Fixed by bounding each fragment with the end of its own `mdat` (`MVLiveIndex.FragLimitByIndex`), capping the unknown-end frontier at the parsed bytes (`ParsedBytes`), and a 64 KiB safety margin; covered by `TestMVLiveIndex_FragLimitFromMdatBeforeNextMoof`.
- **Stripped open-source ICU libs failing to load** (`__register_atfork`). `drm/build-icu-ndk.sh` now retargets the import to `pthread_atfork` after linking.

## Planned

- Wire the vendored wrapper's login/2FA handlers to the DRM auth callback so a missing Android session can be created from the engine.
- Surface lease-recovery state (`Running`, `Scheduled`, `Refreshing`, `Failed`) in `/drm/status` and the SSE `drm` event.
- Progressive MV (itun) decryption through the in-process library.
- Split `server` handlers into focused packages behind small interfaces so individual areas (library, export, catalog) can be embedded without the rest.
- Provide a real recommendations client example that renders artwork and deep links.
- CI: build, `go test ./...` in all modules, `sdk/archtest`, `gofmt` check.
