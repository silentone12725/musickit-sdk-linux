// Package aacstream is the AAC / music-video streaming and key-handling
// pipeline: parallel HLS segment download with a persistent LRU cache, content
// key acquisition (AcquireKey), fragment-by-fragment CBCS/CENC decryption into
// an io.Writer, the encrypted MV cache, and audio-stripping passthrough.
//
// Import boundary: only package fairplay (and the server binary) may import
// this package; sdk/archtest enforces it. Everything else reaches decryption
// through fairplay's pipeline.Decryptor interface.
package aacstream
