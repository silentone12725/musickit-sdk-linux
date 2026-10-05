# mv-player

A browser client for the engine's segmented music-video endpoints, with a mock engine so it runs
without an Apple account.

| File | What it is |
|---|---|
| `vseg-player.js` | The client: `VsegPlayer` implements the whole vseg protocol (setup, fetch loop, seeking, end of stream). About 200 lines, no dependencies; copy it into your own player |
| `index.html` | A small page around it: engine URL, music-video id, Play/Stop, a log |
| `serve.mjs` | Serves this directory on `http://127.0.0.1:8080` |
| `mock-engine.mjs` | A stand-in engine serving an ffmpeg-made test pattern in both seek flavours |

## Run against the mock

Needs Node 20+ and `ffmpeg`.

```sh
node mock-engine.mjs   # http://127.0.0.1:20125
node serve.mjs         # http://127.0.0.1:8080
```

Open <http://127.0.0.1:8080>, set the engine URL to `http://127.0.0.1:20125`, enter any id, press Play, then
drag the seek bar past the buffered range and watch the log. Environment switches for the mock:

- `MOCK_DIRECT=0` never offers the fragment-level path, so seeks use the FFmpeg flavour;
- `MOCK_BAD_DIRECT=1` offers it with a wrong `tsOffset`, so the client's probe rejects it and falls back.

## Run against the real engine

1. Start the engine and sign in (see [../docs/drm.md](../docs/drm.md)).
2. The engine speaks HTTPS with a per-run self-signed certificate: open `https://127.0.0.1:20025/api/v1/status`
   once in the same browser and accept it.
3. `node serve.mjs`, open <http://127.0.0.1:8080>, keep the default engine URL, enter a music-video id.

The engine only answers browsers whose origin is `https://music.apple.com` or a loopback origin, which is why the
page is served from `127.0.0.1`.

The page plays the picture only. The soundtrack is a separate stream (`/playback/{id}/audio`) that a real player
keeps in step with the video clock.

## The seek protocol

`GET /playback/{id}/vseg/seek?t=<seconds>` answers `{n, t, direct?, tsOffset?, reinit?}`. A client must:

1. **Drop** what is buffered and stop the fetch loop.
2. **Place** the new fragments on its timeline:
   - `tsOffset` present (`direct: true`): set `sourceBuffer.timestampOffset = tsOffset`. The fragments keep the
     stream's raw timeline, which sits 10 s ahead of the playlist's, so the offset is negative;
   - otherwise `n === 0`: timestamps restart at 0, so offset by `t`;
   - otherwise (`n > 0`, served from the from-0 producer): already absolute, no offset.
3. **Re-initialise** if `reinit` is true: fetch `/vseg/init` and append it before any fragment. The engine alternates
   between FFmpeg's init segment and the original one.
4. **Fetch** from `/vseg/seg/{n}` until it answers 404, then call `endOfStream()`.

Fragment-level seeks are validated before use: the first one per codec is appended to a throwaway `MediaSource`,
because a rejected append permanently kills a media element. If Chrome refuses it or places it away from `t`,
the client asks again with `&direct=0`, which always returns the FFmpeg flavour, and remembers the verdict.
See [../docs/architecture.md](../docs/architecture.md#mv-seeking) for the engine side.
