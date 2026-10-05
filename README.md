<div align="center">

# musickit-sdk-linux

A local engine and HTTP API that gives Linux applications Apple Music playback with real FairPlay DRM, lossless audio, lyrics, library access and downloads.

</div>

> [!IMPORTANT]
> **Disclaimer.** This project is not affiliated with, authorized by, or endorsed by Apple Inc. "Apple Music" and related marks belong to Apple Inc. It is an independent implementation for personal use and requires an active Apple Music subscription. Review Apple's terms before using it.

## What it is

`musickit-sdk-linux` is a Go engine that runs next to your application and talks to Apple Music on its behalf. Your application (a desktop player, a TUI, a script) speaks plain HTTPS/JSON to `127.0.0.1`; the engine handles everything that is hard to do from a UI process:

- **FairPlay DRM** — key exchange and sample decryption, through an in-process DRM client (`drm/`).
- **Streaming** — HLS parsing, segment download, CBCS decryption, ALAC / AAC / Dolby Atmos / music video.
- **Playback** — optional in-process libvlc output for gapless lossless audio.
- **Catalog & library** — album, playlist and artist detail, encrypted local library cache, lyrics, metadata, artwork.
- **Personalisation** — recommendations, heavy rotation and recently played feeds (see [docs/recommendations.md](docs/recommendations.md); these are provided as ready-to-use but unused endpoints).
- **Prefetch & cache** — a scheduler warms upcoming tracks; persistent disk and segment caches.
- **Export** — queue-based track export with tagging.
- **Events** — a Server-Sent Events channel for DRM state, playback and export progress.

## Quick start

```sh
# 1. Build the DRM client and the engine (see docs/building.md for prerequisites)
make -C drm libdrm_client.so
scripts/build-engine.sh dist

# 2. Run it
cd <directory containing drm/>   # or set drm-binary-path in config.yaml
./dist/musickit-engine --api 20025
```

```sh
# 3. Talk to it (self-signed loopback certificate, hence -k)
curl -sk https://127.0.0.1:20025/api/v1/status
curl -sk https://127.0.0.1:20025/api/v1/capabilities
```

The engine only accepts requests whose `Host` is a loopback address and whose browser `Origin`, if any, is loopback or `https://music.apple.com`. See [docs/security.md](docs/security.md).

## Documentation

| Document | Contents |
|---|---|
| [docs/architecture.md](docs/architecture.md) | Components, data flow, package map |
| [docs/api.md](docs/api.md) | Every HTTP endpoint, request and response shapes |
| [docs/recommendations.md](docs/recommendations.md) | The personalised feed endpoints |
| [docs/drm.md](docs/drm.md) | The DRM client, FairPlay flow, lease recovery, runtime layout |
| [docs/building.md](docs/building.md) | Prerequisites and build steps |
| [docs/configuration.md](docs/configuration.md) | `config.yaml`, flags, environment variables, paths |
| [docs/embedding.md](docs/embedding.md) | Using the engine from your own application (curl, JavaScript, Go) |
| [docs/security.md](docs/security.md) | Threat model, hardening, known issues |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Known issues and planned work |
| [NOTICE.md](NOTICE.md) | Third-party code and credits |

## Repository layout

The SDK is deliberately split into parts that can be used, built and replaced independently.

```
musickit-sdk-linux/
├── drm/        C library (libdrm_client.so): FairPlay session, key exchange, decrypt
├── sdk/        Go module: reusable libraries — media, pipeline, playback, hls, fairplay,
│               drm, ampapi, library, export, prefetch, vlc, …  (no HTTP server)
├── server/     Go module: the HTTP API as an importable package + cmd/musickit-engine
├── examples/   Go module: runnable examples (SDK-only Go program, curl and Node clients)
├── api/        openapi.json
├── scripts/    build helpers
└── docs/       documentation
```

`sdk` never imports `server`, so you can use the libraries without the engine, or write a different front end. Details in [docs/architecture.md](docs/architecture.md); each part has its own README ([sdk](sdk/README.md), [server](server/README.md), [drm](drm/README.md), [examples](examples/README.md)).

## Status

Linux x86_64. ALAC/AAC playback and the DRM client are in daily use. Music-video playback and some endpoints are still evolving; each endpoint's stability is noted in [docs/api.md](docs/api.md).

## License

MIT — see [LICENSE](LICENSE). Third-party components keep their own licences; see [NOTICE.md](NOTICE.md).
