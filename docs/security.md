# Security

## Threat model

The engine is a **local service**. By default it has no credentials of its own: access is limited to the user running it (see the per-user check below) and web pages are kept from driving it. Set `MUSICKIT_API_TOKEN` to add a shared secret on top.

| Control | Where | Purpose |
|---|---|---|
| Loopback bind (`127.0.0.1`) and TLS with a per-run self-signed certificate | `server/tlslocal.go` | Keep the API off the network |
| Per-user check: connections from any uid other than the engine's (or root) are closed | `server/peercred.go` | Loopback is shared by every account on the machine; the client's uid is read from `/proc/net/tcp{,6}`. Disabled with a warning where `/proc` is unavailable, or deliberately with `MUSICKIT_ALLOW_OTHER_USERS=1` |
| Optional shared secret (`MUSICKIT_API_TOKEN`) | `requireToken` | For embedders: `Authorization: Bearer`, `X-Api-Token` or `?access_token=` (for media elements). CORS preflights pass |
| Host check: only `127.0.0.1`, `localhost`, `::1` | `corsPreflightHandler` | DNS-rebinding defence |
| Origin allowlist: `https://music.apple.com` or an exact loopback origin | `corsPreflightHandler` | CSRF defence; cross-site "simple" POSTs skip CORS preflight, so this rejects them outright. Never loosen to prefix matching |
| Parsed-hostname comparisons | `isLoopbackHost`, `isAllowedOrigin` | Prevents `localhost.attacker.example` style bypasses |
| `X-Content-Type-Options: nosniff` | middleware | Catalog text must not be sniffed into executable types |
| Path-parameter validation | `catalogParams`, `typesFilterRe` | Values are interpolated into Apple API URLs together with bearer tokens |
| Request-supplied tool paths: name must be `ffmpeg`, `ffmpeg-*`, `vlc`, `vlc-*` or `cvlc` (also after resolving symlinks), and the file must be an executable regular file owned by root or the engine's user, not writable by others, in a directory nobody else can write | `export.validToolPath`, `checkToolExecutable` | No arbitrary command execution |
| Export output confined to allowed roots (the user's home, or `MUSICKIT_EXPORT_ROOTS`), compared after resolving symlinks | `export.Options.OutputRoots` | A request body from a web page names the directory files are written to |
| pprof only with `MUSICKIT_DEBUG=1` | `NewAPIServer` | Heap dumps expose key material |
| Single-instance lock | `engine-session.lock` | One engine owns a DRM session |

Do not expose the port beyond loopback (reverse proxy, port forward, container publish) without adding your own authentication.

## Secrets

- `<drm dir>/files/` contains the Apple session (`mpl_db`, `MUSIC_TOKEN`). Treat it as a password; keep it out of version control.
- Tokens supplied through `POST /library/token` or `POST /playback` stay in process memory.
- The encrypted library cache key lives next to the cache, so the encryption guards against accidental disclosure, not against someone who can read the directory; protect `~/.cache/musickit-sdk-linux/`. Keys are created exclusively and never regenerated over an unreadable file.
- The DRM base directory (`<drm dir>/files/`) is restricted to mode 0700 at start, and the cookie jar inside it is written atomically with mode 0600.

## Known issues to address before distributing builds

- **Widevine device identity is not shipped.** AAC/MV key requests need `device_private_key` and `device_client_id_blob` in `$MUSICKIT_WIDEVINE_DIR` (default `~/.config/musickit-sdk-linux/widevine`). Without them those requests fail with `ErrNoDeviceIdentity`. Keep the directory mode 0700. Earlier commits of this repository contained a default identity; rotate it before making the repository public.
- **Vendored wrapper licence.** `drm/native/` derives from an MIT-licensed wrapper (see [NOTICE.md](../NOTICE.md) and `drm/native/LICENSE.wrapper`); keep the licence files with any redistribution. The `zhaarey/wrapper` lineage it was forked from declares no licence.
- **TLS verification in examples.** `examples/http/client.mjs` disables certificate verification for the loopback connection only; do not copy that into code that talks to other hosts.

## Reporting

Open an issue marked *security* without including tokens, session files or heap dumps.
