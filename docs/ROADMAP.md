# Roadmap and known issues

Work that is queued, in rough priority order.

## Known issues

1. **Music-video playback stops after about a second** in the segmented-video path: the browser reports `CHUNK_DEMUXER_ERROR_APPEND_FAILED: Failed to prepare video sample for decode` while the first fragments decode fine. The video is decrypted by `sdk/aacstream` (Widevine key) and stream-copied by FFmpeg in `server/handlers_vseg.go`; the lossless audio path is unaffected. Next step: dump the first failing fragment, check subsample/pattern handling after decryption, and compare with the audio track.
2. **Stripped open-source ICU build** only loads after patching its import `__register_atfork` to `pthread_atfork`. Build the patch into `drm/build-icu-ndk.sh` (post-link step inside the container) or keep using Apple's ICU libraries.
3. **Embedded Android libraries break FairPlay** (see [drm.md](drm.md)). Decide whether to remove the embedding path or keep it as diagnostics only; shipping requires bundling `rootfs/system/lib64`.
4. **Embedded Widevine identity** in `sdk/aacstream/cdm/consts.go` (see [security.md](security.md)).

## Planned

- Wire the vendored wrapper's login/2FA handlers to the DRM auth callback so a missing Android session can be created from the engine.
- Surface lease-recovery state (`Running`, `Scheduled`, `Refreshing`, `Failed`) in `/drm/status` and the SSE `drm` event.
- Progressive MV (itun) decryption through the in-process library.
- Split `server` handlers into focused packages behind small interfaces so individual areas (library, export, catalog) can be embedded without the rest.
- Provide a real recommendations client example that renders artwork and deep links.
- CI: build, `go test ./...` in all modules, `sdk/archtest`, `gofmt` check.
