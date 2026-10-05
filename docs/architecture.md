# Architecture

The SDK is three independent parts that you can use separately.

```
musickit-sdk-linux/
├── drm/        C library  libdrm_client.so      FairPlay session, key exchange, decrypt
├── sdk/        Go module  …/sdk                 reusable libraries (no HTTP server)
├── server/     Go module  …/server              HTTP API over the sdk + the engine binary
├── examples/   Go module  …/examples            runnable examples
├── api/        openapi.json                     machine-readable API summary
└── scripts/    build helpers
```

`go.work` ties the three Go modules together for development; each has its own `go.mod`. Dependencies point one way:

```
examples ─┐
          ├─► sdk ─► (cgo) ─► drm/libdrm_client.so ─► Android FairPlay libs (via libhybris)
server  ──┘
```

`sdk` never imports `server`; `server` is just one possible front end. You can build a different one (a TUI, gRPC, an MPRIS bridge) on the same libraries.

## sdk — packages

Flat, one directory per concern. Dependency rules are enforced by `sdk/archtest`.

| Package | Role |
|---|---|
| `media` | Provider-agnostic interfaces (`Provider`, `Session`, `Track`) so the engine is not tied to Apple Music |
| `pipeline` | Core types: `Source`, `Stage`, `Decryptor`, `Stream`, and `Run` |
| `playback` | Session lifecycle: open → stream, TTL, reuse, release |
| `apple` | `media.Provider` implementation for Apple Music (catalog → HLS → pipeline) |
| `hls` | HLS playlist parsing (master and media), seek helpers |
| `fairplay` | The only DRM adapter: key acquisition and CBCS decrypt sources for the pipeline |
| `drm` | DRM manager and backends (state machine, crash recovery, native in-process backend over cgo) |
| `aacstream`, `alacstream` | Segment download, fragment decryption and ALAC helpers used by `fairplay` |
| `diskcache` | Download-then-serve cache for lossless tracks |
| `prefetch` | Generation-numbered scheduler that warms upcoming tracks |
| `library` | AES-256-GCM encrypted library cache |
| `export` | Job queue, remux, tagging and templates for exports |
| `ampapi` | Apple Music catalog and personalised-feed client |
| `lyrics` | Lyrics fetch and format conversion |
| `vlc` | In-process libvlc output (cgo) |
| `ffmpeg`, `autoeq`, `cavern` | Optional audio stages: transcode, headphone EQ, Atmos binaural rendering |
| `ring`, `tracer`, `config`, `mvlabel` | Small utilities |
| `internal/*` | Packages private to the sdk module |

Boundary rules (see `sdk/archtest`): `aacstream` is imported only by `fairplay`; `drm` and `fairplay` do not import each other (the `CBCSDialer` interface lives in `fairplay`); `media` has no Apple-specific dependency; `sdk` never imports `server`.

## server — HTTP layer

`server` is an importable package (`NewAPIServer`, `ServerConfig`) plus `cmd/musickit-engine`, a thin `main` that loads `config.yaml`, parses `--api <port>` and starts it. Handlers are grouped by file: `handlers_playback.go`, `handlers_drm.go`, `handlers_catalog.go`, `handlers_recommendations.go`, `handlers_library.go`, `handlers_export.go`, `handlers_vlc.go`, and so on. Cross-cutting pieces: the loopback TLS certificate (`tlslocal.go`), the host/origin guard (`corsPreflightHandler`), the SSE event bus, and the DRM lifecycle watcher.

To add an endpoint: write a handler method on `APIServer`, register it in `NewAPIServer`'s route block, document it in `docs/api.md` and `api/openapi.json`, and add a test next to it (see `handlers_recommendations_test.go` for the pattern: a fake upstream plus `httptest`).

## Data flow: playing a lossless track

1. A client `POST`s `/playback` with an `assetId`. `playback.Manager` asks the `apple` provider for the track; the provider resolves the HLS playlist and returns a `media.Session`.
2. The client (or `/vlc/load`) requests the audio. `playback.Manager.Stream` builds a `pipeline` from a `fairplay` CBCS source.
3. `fairplay` downloads fragments and, per key, asks the DRM backend for a decrypt context (`DialCBCS`). The `drm` package opens it through `libdrm_client.so`, which performs the FairPlay key exchange with Apple and returns a handle.
4. Decrypted fragments are written to the disk cache and served with range support; libvlc reads them in-process.
5. The prefetch scheduler, driven by `PUT /playback/context`, warms the next tracks the same way.

## Extension points

- **A new media provider:** implement `media.Provider`; nothing else in the pipeline changes.
- **A new pipeline stage:** implement `pipeline.Stage` (see `sdk/ffmpeg` for audio stages).
- **A different front end:** import `sdk/...` packages directly; `examples/go-recommendations` does this for the catalog client.
- **A different DRM backend:** implement `drm.DRMBackend`; the manager handles state, recovery and authentication around it.
