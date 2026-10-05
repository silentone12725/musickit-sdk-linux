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

An AES-256-GCM encrypted local cache of the listener's library. The engine does not fetch the library itself: the client (which holds the user's MusicKit session) downloads it and hands it over with `/library/ingest`.

| Method | Path | Description |
|---|---|---|
| POST | `/library/token` | Hand the engine the listener's tokens. Body: `{"musicUserToken": "…", "developerToken": "…"}` (developer token optional; max 16 KB). Returns `{"ok": true}`. Needed for `/recommendations*` and live library fetches |
| POST | `/library/ingest` | Replace the cache with a payload (max 100 MB). Returns `{"songs","albums","playlists","syncedAt"}` |
| POST | `/library/sync` | **Deprecated**: always `410 Gone`. Use `/library/ingest` |
| GET | `/library/status` | `{"songs","albums","playlists","syncedAt","needsSync"}`; `503` if the store is unavailable |
| GET | `/library/playlists` | `{"playlists": [...]}` from the cache |
| GET | `/library/playlists/{id}/tracks` | `{"tracks": [{"lid","cid"}], "cached": bool}`. `lid` is the library song ID, `cid` the catalog ID. On a cache miss with tokens available, the engine fetches live and caches the result; with no tokens it returns an empty list and `cached:false` |
| GET | `/library/albums/{id}/tracks` | `{"tracks": [{"lid","cid"}], "cached": true}` from the local database. `l.` is prefixed to the ID if missing |

`POST /library/ingest` body (items follow Apple Music API resource objects):

```json
{
  "songs": [ { "id": "i.…", "type": "library-songs", "attributes": { } } ],
  "albums": [ ],
  "playlists": [ ],
  "playlistTracks": { "<playlistId>": [ ] },
  "revision": "<opaque token for delta sync>"
}
```

## Cache

| Method | Path | Description |
|---|---|---|
| GET | `/cache/stats` | `{"persistent": {"available","sizeBytes","limitBytes","ttlDays","count"}, "prewarm": {"entries","sizeBytes","limitBytes"}}`. Defaults shown when no limit is set: 500 MB persistent, 1 GB prewarm |
| PUT | `/cache/config` | Body (max 4 KB): `{"prewarmLimitMB": N, "persistLimitMB": N, "persistTTLDays": N}`; omitted fields are unchanged. Returns the effective config |
| DELETE | `/cache/playback?what=` | `persistent` (cached songs), `prewarm`, `segments`, or empty for all. `204` |
| GET | `/cache/mv` | `{"enabled","maxBytes","sizeBytes","quality"}` for the music-video cache |
| PUT | `/cache/mv` | Body (max 4 KB): `{"enabled": bool, "maxBytes": N}`, both optional. Returns the same shape as `GET` |
| DELETE | `/cache/mv` | Clear the music-video cache. `204` |

## Export

Exports are queued jobs; progress is also pushed as `export` SSE events.

| Method | Path | Description |
|---|---|---|
| POST | `/export` | Enqueue a job (max 64 KB). `202` with the job. `401` if no developer token or media token is available (supply `token` and `mut` in the body, or start playback first). `400` for an invalid request |
| GET | `/export` | List all jobs |
| GET | `/export/{id}` | One job; `404` if unknown |
| DELETE | `/export/{id}` | Cancel; `204`, or `404` |
| POST | `/export/{id}/retry` | Retry a `failed` or `cancelled` job; `202` with the job, `400` if not retryable, `404`, `401` without tokens |
| POST | `/export/{id}/priority` | Body: `{"priority": N}`. Higher runs sooner, FIFO among equals. `200` with the job, `404` unknown, `409` if already running or finished |

`POST /export` body:

```json
{
  "assetId": "1488408568",
  "storefront": "us",
  "token": "<optional developer JWT>",
  "mut": "<optional Music-User-Token>",
  "language": "en-US",
  "capabilities": { "lossless": true, "atmos": false, "video": false,
                    "playlist": false, "libraryPlaylist": false },
  "mvMaxHeight": 0,
  "outputDir": "/home/me/Music",
  "filenameTemplate": "{album_artist}/{album}/{track_number:02d} - {title}",
  "options": {
    "embedArtwork": true, "artworkSize": 3000,
    "embedLyrics": true, "lrcFormat": "lrc", "lrcType": "lyrics", "saveLrcSidecar": false,
    "overwritePolicy": "skip",
    "convertToFlac": false, "keepOriginal": false,
    "explicitChoice": "[E]", "cleanChoice": "[C]", "masterChoice": "[M]"
  },
  "hintTitle": "", "hintArtist": "", "hintArtwork": "",
  "priority": 0
}
```

- `capabilities.playlist` expands a playlist into one job per track (children inherit `priority`); `libraryPlaylist` uses the library API for `p.…` IDs.
- `overwritePolicy`: `skip` (default), `overwrite` or `rename`. `lrcFormat`: `lrc` or `ttml`. `lrcType`: `lyrics` or `syllable-lyrics`.
- Filename template variables: `{title}`, `{artist}`, `{album_artist}`, `{album}`, `{track_number}`, `{track_number:02d}`, `{disc_number}`, `{year}`, `{genre}`, `{codec}`, `{ext}`.
- `hint*` fields only let a UI show the job row before the catalog lookup finishes.

A job looks like:

```json
{
  "jobId": "…", "assetId": "…", "phase": "downloading", "percent": 42,
  "queuePos": 7, "queueIndex": 0, "priority": 0,
  "title": "…", "artistName": "…", "artworkUrl": "…",
  "bytesDone": 0, "bytesTotal": 0, "output": "", "error": "",
  "source": "network", "limitBps": null, "throttled": false,
  "createdAt": "…", "updatedAt": "…"
}
```

`phase` is one of `queued`, `resolving`, `downloading`, `tagging`, `moving`, `done`, `failed`, `cancelled`. `source` is `cache`, `playback` or `network`: where the bytes came from.

## Errors

| Status | Meaning |
|---|---|
| 400 | Malformed input or failed validation |
| 401 | Missing credentials/tokens (for example no Music-User-Token) |
| 403 | Forbidden host/origin, or Apple rejected the token |
| 502 | Apple's service failed or returned an unexpected status |
| 503 | A required backend (DRM, VLC) is unavailable |
