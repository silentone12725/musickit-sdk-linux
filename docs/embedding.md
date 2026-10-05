# Using the engine from your application

There are two ways to build on the SDK. Pick per component; they mix freely.

1. **Run the engine and talk HTTP.** Any language works. Best for UIs and scripts.
2. **Import the Go libraries.** Use `sdk/...` packages directly (catalog client, HLS parser, pipeline, DRM manager) in your own Go program, or embed `server` and add your own routes.

## Talking to the engine

Start it (`musickit-engine --api 20025`), then use `https://127.0.0.1:20025/api/v1/...`. The certificate is self-signed and regenerated per run; trust it for that connection only. Requests must have a loopback `Host`; browser `Origin`, if present, must be loopback or `https://music.apple.com` (see [security.md](security.md)).

Typical flow for a player:

1. `GET /status`, `GET /capabilities`, `GET /drm/status` — wait for DRM `RUNNING` (or watch `GET /events`).
2. `POST /library/token` with the listener's Music-User-Token (enables personalised feeds and library).
3. `PUT /playback/context` as the queue changes, so upcoming tracks are prefetched.
4. `POST /playback` for the track → `GET /playback/{id}/audio` (or `POST /vlc/load` to play in-process).
5. `DELETE /playback/{id}` when done.

### curl

```sh
curl -sk https://127.0.0.1:20025/api/v1/capabilities
curl -sk -X POST https://127.0.0.1:20025/api/v1/playback \
  -H 'Content-Type: application/json' \
  -d '{"assetId":"1488408568","storefront":"us","capabilities":{"lossless":true}}'
```

### JavaScript

`examples/http/client.mjs` (Node 20+) reads status and capabilities and tails the event stream. In a browser, call from a page served from a loopback origin.

### Server-Sent Events

```js
const es = new EventSource('https://127.0.0.1:20025/api/v1/events');
es.addEventListener('drm', (e) => console.log(JSON.parse(e.data)));
```

## Using the libraries

```go
import "github.com/silentone12725/musickit-sdk-linux/sdk/ampapi"

feed, err := ampapi.GetRecommendations(ctx, ampapi.KindRecommendations, devToken, userToken,
    ampapi.RecommendationsOptions{Language: "en-US", Limit: 10})
```

`examples/go-recommendations` is a complete program. In your own module:

```
require github.com/silentone12725/musickit-sdk-linux/sdk v0.0.0
replace github.com/silentone12725/musickit-sdk-linux/sdk => /path/to/musickit-sdk-linux/sdk
```

### Embedding the server

`server.NewAPIServer(port, server.ServerConfig{...})` returns an `*APIServer` with `Start()` and `Stop()`. To add routes, add a handler file to `server/` following `handlers_recommendations.go` and register it in the route block of `NewAPIServer`.

### Adding a provider or stage

Implement `media.Provider` for another catalog, or `pipeline.Stage` for an audio transform. Neither requires touching `server`.

## Adding an endpoint checklist

1. Handler + validation in `server/handlers_<area>.go`.
2. Route in `NewAPIServer`.
3. Test with `httptest` (fake upstream where needed).
4. Entry in `docs/api.md` and `api/openapi.json`.
