# Personalised feeds (recommendations)

Three endpoints expose the listener's personalised feeds. They are **dormant**: complete and covered by tests, but not used by any bundled frontend, so other applications can adopt them without changing the engine.

| Endpoint | Apple Music API route | Content |
|---|---|---|
| `GET /api/v1/recommendations` | `/v1/me/recommendations` | "Listen Now" feed |
| `GET /api/v1/recommendations/heavy-rotation` | `/v1/me/history/heavy-rotation` | Most-played content |
| `GET /api/v1/recommendations/recently-played` | `/v1/me/recent/played` | Recently played albums, playlists and stations |

## Query parameters

| Name | Rule | Description |
|---|---|---|
| `limit` | integer 1-50 | Page size (Apple's default if omitted) |
| `offset` | integer 0-10000 | Page offset |
| `types` | comma-separated, lowercase letters/dashes, at most 10 | Filter resource types, e.g. `playlists,albums` |

The language is taken from the `Accept-Language` header.

## Authentication

These routes act on behalf of a listener, so a **Music-User-Token** is required in addition to the developer token. Provide it once:

```sh
curl -sk -X POST https://127.0.0.1:20025/api/v1/library/token \
  -H 'Content-Type: application/json' \
  -d '{"musicUserToken":"<token>","developerToken":"<optional JWT>"}'
```

If no token has been pushed, the engine falls back to the token in the DRM session. With none available the endpoint returns `401`.

## Responses

The body is Apple's JSON, passed through unchanged, so every field Apple adds is visible without an SDK update. Errors:

| Status | Cause |
|---|---|
| 400 | Invalid `limit`, `offset` or `types` |
| 401 | No Music-User-Token available, or Apple returned 401 |
| 403 | Apple returned 403 (for example no active subscription) |
| 502 | Network failure or any other upstream status |

## Example

```sh
curl -sk 'https://127.0.0.1:20025/api/v1/recommendations?limit=10&types=playlists' | jq '.data[].attributes.title'
```

```js
const r = await fetch('https://127.0.0.1:20025/api/v1/recommendations/heavy-rotation?limit=20');
const { data } = await r.json();
```

## Implementation notes

- Client: `engine/utils/ampapi/recommendations.go` (`GetRecommendations`, `RecommendationKind`, `StatusError`). The base URL is the variable `ampapi.AMPBaseURL` so tests can substitute a fake server.
- Handlers and validation: `engine/cmd/handlers_recommendations.go`; routes are registered next to the catalog routes in `engine/cmd/apiserver.go`.
- Tests: `engine/cmd/handlers_recommendations_test.go` cover header forwarding, path mapping, validation, auth and upstream error mapping.
- To add another personalised feed, add a `RecommendationKind` constant and path in `recommendations.go` and register one more route.
