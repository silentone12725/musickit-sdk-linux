# Security

## Threat model

The engine is a **local service**. It is unauthenticated, so everything depends on only the local user being able to reach it and on web pages not being able to drive it.

| Control | Where | Purpose |
|---|---|---|
| Loopback bind (`127.0.0.1`) and TLS with a per-run self-signed certificate | `server/tlslocal.go` | Keep the API off the network |
| Host check: only `127.0.0.1`, `localhost`, `::1` | `corsPreflightHandler` | DNS-rebinding defence |
| Origin allowlist: `https://music.apple.com` or an exact loopback origin | `corsPreflightHandler` | CSRF defence; cross-site "simple" POSTs skip CORS preflight, so this rejects them outright. Never loosen to prefix matching |
| Parsed-hostname comparisons | `isLoopbackHost`, `isAllowedOrigin` | Prevents `localhost.attacker.example` style bypasses |
| `X-Content-Type-Options: nosniff` | middleware | Catalog text must not be sniffed into executable types |
| Path-parameter validation | `catalogParams`, `typesFilterRe` | Values are interpolated into Apple API URLs together with bearer tokens |
| Request-supplied tool paths accepted only when the file name is `ffmpeg*`, `vlc` or `cvlc` | export handlers | No arbitrary command execution |
| pprof only with `MUSICKIT_DEBUG=1` | `NewAPIServer` | Heap dumps expose key material |
| Single-instance lock | `engine-session.lock` | One engine owns a DRM session |

Do not expose the port beyond loopback (reverse proxy, port forward, container publish) without adding your own authentication.

## Secrets

- `<drm dir>/files/` contains the Apple session (`mpl_db`, `MUSIC_TOKEN`). Treat it as a password; keep it out of version control.
- Tokens supplied through `POST /library/token` or `POST /playback` stay in process memory.
- The encrypted library cache key lives next to the cache; protect `~/.cache/musickit-sdk-linux/`.

## Known issues to address before distributing builds

- **Widevine device identity is not shipped.** AAC/MV key requests need `device_private_key` and `device_client_id_blob` in `$MUSICKIT_WIDEVINE_DIR` (default `~/.config/musickit-sdk-linux/widevine`). Without them those requests fail with `ErrNoDeviceIdentity`. Keep the directory mode 0700. Earlier commits of this repository contained a default identity; rotate it before making the repository public.
- **Proprietary libraries are in the repository.** `drm/rootfs/system/lib64` contains Apple's and Android's binaries. Publishing the repository redistributes them; get clearance first, or remove that directory (`git rm -r drm/rootfs/system/lib64`, restore the ignore rule, and have users run `scripts/install-android-libs.sh`). Never commit `drm/rootfs/data`.
- **Vendored wrapper licence.** `drm/native/` is code from a third-party reference wrapper. Confirm its licence terms before redistributing (see [NOTICE.md](../NOTICE.md)).
- **TLS verification in examples.** `examples/http/client.mjs` disables certificate verification for the loopback connection only; do not copy that into code that talks to other hosts.

## Reporting

Open an issue marked *security* without including tokens, session files or heap dumps.
