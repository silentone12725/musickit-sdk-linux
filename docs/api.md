# HTTP API reference

Base URL: `https://127.0.0.1:<port>/api/v1` (the port is the `--api` flag). The certificate is generated per run for loopback use; clients must either trust it for that connection (`curl -k`) or pin it.

Conventions:

- Bodies are JSON unless noted. Errors are plain-text bodies with an appropriate status code.
- `{id}` path values and `sf` (storefront) query values are validated against `^[A-Za-z0-9._-]{1,64}$`; anything else returns `400`.
- Language comes from the `Accept-Language` header (default `en-US`).
- Stability: **stable** (relied on by the reference frontend), **evolving**, **dormant** (implemented and tested, not used by any bundled frontend).

## System

| Method | Path | Description | Stability |
|---|---|---|---|
| GET | `/status` | Health check | stable |
| GET | `/capabilities` | Feature flags: lossless, Atmos, CBCS, VLC availability | stable |
| GET | `/events` | Server-Sent Events channel (see below) | stable |
| GET | `/tools` | Which helper tools (ffmpeg, MP4Box, VLC) were found | evolving |
| GET | `/metrics` | Internal latency and cache metrics | evolving |
| GET | `/debug/runtime` | Runtime stats | evolving |
| GET | `/debug/pprof/*` | Go pprof; only registered when `MUSICKIT_DEBUG=1` (heap dumps can contain key material) | debug |

### Events (SSE)

`GET /events` streams `id: N`, `event: <type>`, `data: <json>`. Reconnect with `Last-Event-ID` to replay missed events from the replay buffer. Event types: `engine.snapshot` (sent first), `drm`, `drm.refresh_due`, `playback.created`, `playback.deleted`, `export`.

## DRM session

| Method | Path | Body | Description |
|---|---|---|---|
| GET | `/drm/status` | — | DRM snapshot (process, manager, authentication, FairPlay, session, recovery state), selected backend, session age and TTL |
| POST | `/drm/authenticate` | `{"email","password"}` | Start a login (max 4 KB). `503` when the DRM backend is unavailable |
| POST | `/drm/challenge` | `{"reply"}` | Answer a pending 2FA challenge |
| POST | `/drm/logout` | — | Sign out |
| DELETE | `/drm/session` | — | Drop the stored session |

## Playback

| Method | Path | Description |
|---|---|---|
| POST | `/playback` | Create a session. Body below. Returns session metadata (codec, sample rate, bit depth, duration, artwork URL, `expiresIn`) |
| GET | `/playback/{id}/audio` | Stream audio (ALAC/AAC/Atmos). Supports `Range`; ALAC is served from the disk cache |
| GET | `/playback/{id}/video` | Music video stream |
| GET | `/playback/{id}/video-es`, `video-raw`, `video-dl`, `video-dl-info` | Alternative music-video delivery modes (evolving) |
| GET | `/playback/{id}/vseg/manifest`, `init`, `seg/{n}`, `seek` | Segmented video for MSE players (evolving) |
| DELETE | `/playback/{id}/vseg` | Stop a segmented-video producer |
| POST | `/playback/{id}/precache` | Warm a session's audio into the cache |
| DELETE | `/playback/{id}` | Release a session |
| PUT | `/playback/context` | Tell the engine the current queue so the prefetcher can warm upcoming tracks. Returns `202 {"jobId"}` |
| GET / DELETE | `/jobs/{id}` | Inspect or cancel a cache-warm job |

`POST /playback` body:

```json
{
  "assetId": "1488408568",
  "storefront": "us",
  "token": "<optional developer JWT override>",
  "mediaUserToken": "<optional Music-User-Token override>",
  "capabilities": { "lossless": true, "atmos": false, "video": false },
  "mvMaxHeight": 0
}
```

`assetId` is required (max body 64 KB). If `token`/`mediaUserToken` are omitted the engine uses its cached tokens and the DRM session's.

`PUT /playback/context` body (max 512 KB, up to ~50 tracks):

```json
{
  "context": { "type": "album", "id": "…", "reason": "play" },
  "currentIndex": 0,
  "tracks": [ { "id": "…" } ],
  "lossless": true
}
```

## In-process player (libvlc)

Routes are no-ops when libvlc is not installed (check `/capabilities`).

| Method | Path | Body |
|---|---|---|
| POST | `/vlc/load` | `{"sessionId","assetId","startMs"}` |
| POST | `/vlc/pause`, `/vlc/resume`, `/vlc/stop` | — |
| POST | `/vlc/seek` | `{"posMs","sessionId"}` |
| POST | `/vlc/volume` | `{"volume": 0-100}` |
| GET | `/vlc/time` | position / state |

## Metadata, artwork, lyrics

| Method | Path | Description |
|---|---|---|
| GET | `/metadata/{id}?sf=` | Track info and available qualities (`StreamInfo`: codec, sampleRate, bitDepth, bitrate) |
| GET | `/artwork/{id}` | Proxy an artwork image from Apple's CDN |
| GET | `/lyrics/{id}` | Lyrics (word-by-word, translation, pronunciation where available) |
| GET | `/audioanalysis/{id}` | Audio analysis data |

## Catalog

Thin, validated proxies for Apple Music catalog detail. Query: `sf` (storefront).

| Method | Path |
|---|---|
| GET | `/catalog/albums/{id}` |
| GET | `/catalog/playlists/{id}` |
| GET | `/catalog/artists/{id}` (includes albums and top songs) |

## Personalised feeds (dormant)

See [recommendations.md](recommendations.md).

| Method | Path |
|---|---|
| GET | `/recommendations` |
| GET | `/recommendations/heavy-rotation` |
| GET | `/recommendations/recently-played` |

## Library

An AES-256-GCM encrypted local cache of the user's library.

| Method | Path | Description |
|---|---|---|
| POST | `/library/token` | `{"musicUserToken","developerToken"}` — hand the engine the listener's tokens (max 16 KB). Required for `/recommendations*` |
| POST | `/library/sync` | Sync from Apple |
| POST | `/library/ingest` | Ingest a payload produced by a frontend |
| GET | `/library/status` | Cache status |
| GET | `/library/playlists` | Playlists |
| GET | `/library/playlists/{id}/tracks` | Tracks of a library playlist |
| GET | `/library/albums/{id}/tracks` | Tracks of a library album |

## Cache

| Method | Path | Description |
|---|---|---|
| GET | `/cache/stats` | Statistics |
| PUT | `/cache/config` | Change cache limits |
| DELETE | `/cache/playback?what=` | `persistent` (cached songs), `prewarm`, `segments`, or empty for all |
| GET / PUT / DELETE | `/cache/mv` | Music-video segment cache: info, configure, clear |

## Export

| Method | Path | Description |
|---|---|---|
| POST | `/export` | Enqueue a job |
| GET | `/export` | List jobs |
| GET / DELETE | `/export/{id}` | Status / cancel |
| POST | `/export/{id}/retry` | Retry a failed or cancelled job |
| POST | `/export/{id}/priority` | Reprioritise |

## Errors

| Status | Meaning |
|---|---|
| 400 | Malformed input or failed validation |
| 401 | Missing credentials/tokens (for example no Music-User-Token) |
| 403 | Forbidden host/origin, or Apple rejected the token |
| 502 | Apple's service failed or returned an unexpected status |
| 503 | A required backend (DRM, VLC) is unavailable |
