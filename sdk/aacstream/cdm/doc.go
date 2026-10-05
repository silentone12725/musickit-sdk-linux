// Package wv implements the client side of a Widevine license exchange
// (license request/response handling, PSSH parsing and the protobuf types) used
// by sdk/aacstream to obtain content keys for AAC and music-video streams.
//
// NOTE: consts.go embeds a default CDM identity. Applications that
// distribute builds should supply their own through configuration instead.
package wv
