# server

The HTTP API as a Go package (`server`) plus the engine binary (`cmd/musickit-engine`).

```sh
go build ./cmd/musickit-engine          # DRM disabled
../scripts/build-engine.sh ../dist      # with in-process DRM (needs ../drm/libdrm_client.so)
./musickit-engine --api 20025
```

- Handlers are grouped by area in `handlers_*.go`; routes are registered in `NewAPIServer` (`apiserver.go`).
- Security middleware: `corsPreflightHandler`, loopback TLS (`tlslocal.go`).
- `config.yaml.example` documents every configuration key.
- Endpoint reference: [../docs/api.md](../docs/api.md). Add new endpoints following the checklist in [../docs/embedding.md](../docs/embedding.md).

Embed it in your own binary with `server.NewAPIServer(port, server.ServerConfig{...})`.
