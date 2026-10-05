# examples

| Example | Needs | Shows |
|---|---|---|
| `go-recommendations/` | `MUSICKIT_USER_TOKEN` | Using `sdk/ampapi` directly, with no engine running |
| `http/client.sh` | a running engine, `curl` | Status, capabilities, DRM state, creating and streaming a playback session (`ASSET=<id>`) |
| `http/client.mjs` | a running engine, Node 20+ | Reading status and tailing the SSE event stream |
| `mv-player/` | a browser; the engine or `mock-engine.mjs` (Node 20+, `ffmpeg`) | A music-video MSE client: the vseg protocol, seeking, and the fragment-level seek with its fallback |

```sh
cd examples && go run ./go-recommendations
```
