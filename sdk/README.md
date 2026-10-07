# sdk

Reusable Go libraries behind the engine. No HTTP server, no global state beyond caches; import only what you need.

```
require github.com/silentone12725/musickit-sdk-linux/sdk v0.0.0
```

Start with the package map in [../docs/architecture.md](../docs/architecture.md). Every package has a doc comment (`go doc ./...`). Import boundaries are tested by `archtest`:

```sh
go test ./...            # all packages, including archtest
go test ./archtest       # boundaries only
```

Packages that use cgo: `drm` (build tag `widevine_backend`, links `drm/libdrm_client.so`) and `vlc` (libvlc).

Good entry points: `ampapi` (catalog and personalised feeds), `hls` (playlist parsing), `media` + `pipeline` + `playback` (the provider/stream model), `library` (encrypted local cache).
