// Package ampapi is a small client for the Apple Music catalog API
// (amp-api.music.apple.com): developer-token retrieval, album / playlist /
// song / music-video / search lookups, and the personalised feeds in
// recommendations.go (recommendations, heavy rotation, recently played).
//
// Functions take the tokens they need explicitly; the package keeps no global
// credentials. AMPBaseURL can be replaced in tests.
package ampapi
