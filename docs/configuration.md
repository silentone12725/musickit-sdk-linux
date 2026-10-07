# Configuration

## Command line

```
musickit-engine --api <port>
```

`--api` starts the HTTP server on `127.0.0.1:<port>`. Without it the binary prints usage and exits.

## config.yaml

Read from the working directory. If it is missing, built-in defaults are used and a warning is logged. `server/config.yaml.example` lists every key; the ones that matter in API mode:

| Key | Description |
|---|---|
| `storefront` | Two-letter default storefront (default `us`) |
| `media-user-token`, `authorization-token` | Optional token overrides |
| `language` | Default language |
| `drm-binary-path` | Marker path whose directory is the DRM directory (default: `./drm`) |
| `drm-base-dir` | DRM session directory (default: `<drm dir>/files`) |
| `stream-cache-size` | Cache size in MB (0 = unlimited) |
| `max-memory-limit` | Memory limit in MB |
| `export-throttle-floor-kbps` | Minimum export rate while music is playing |
| `alac-max`, `atmos-max`, `aac-type`, `mv-max`, `mv-audio-type` | Quality preferences |

The remaining keys (folder templates, tagging, conversion, lyrics format) apply to the export pipeline.

## Environment variables

| Variable | Effect |
|---|---|
| `MUSICKIT_DISABLE_HIRES=1` | Report `hiRes: false` in `/capabilities` and the DRM snapshot (the account token carries no tier, so hi-res is otherwise assumed once FairPlay is ready) |
| `MUSICKIT_TLS_PINS` | Comma-separated base64 SPKI SHA-256 pins for the DRM client's HTTPS connections; when set, a handshake succeeds only if some certificate in the served chain matches |
| `MUSICKIT_WIDEVINE_DIR` | Directory holding `device_private_key` and `device_client_id_blob` (default `~/.config/musickit-sdk-linux/widevine`); required for AAC/MV key requests |
| `MUSICKIT_MV_DIRECT_SEEK=0` | Disable fragment-level MV seeks and always use the FFmpeg seek producer (on by default for H.264 streams) |
| `MUSICKIT_EXPORT_ROOTS` | Colon-separated directories exports may write into (default: the user's home). A request's `outputDir` outside them is refused |
| `MUSICKIT_API_TOKEN` | When set, every API request must carry it (`Authorization: Bearer`, `X-Api-Token` or `?access_token=`) |
| `MUSICKIT_ALLOW_OTHER_USERS=1` | Turn off the per-user connection check (any local account may then use the API) |
| `MUSICKIT_HTTPS_TIMEOUT_SEC` | Connect/read timeout in seconds for the DRM client's HTTPS requests (default 30) |
| `MUSICKIT_DEBUG=1` | Register `/debug/pprof/*` and verbose AAC debug output. Heap dumps can expose key material |
| `MUSICKIT_CAVERN` | Path to `CavernPipeServer` for lossless Atmos binaural rendering |

## Paths

| Path | Content |
|---|---|
| `<drm dir>/files/` | DRM session: `mpl_db/`, `MUSIC_TOKEN`, `STOREFRONT_ID` (sensitive; never commit) |
| `<drm dir>/files/engine-session.lock` | Single-instance lock |
| `~/.cache/musickit-sdk-linux/playback/` | Cached lossless tracks (`{assetId}-alac.m4a`) |
| `~/.cache/musickit-sdk-linux/library.enc`, `library.key` | Encrypted library cache and its key |
| `~/.cache/musickit-sdk-linux/segments` | AAC segment cache |
| `~/.config/musickit-sdk-linux/drm/` | Optional per-user DRM directory (searched for `lib64` and `q.so`) |

Clear caches with `DELETE /api/v1/cache/playback?what=persistent|prewarm|segments`.
