package server

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Without libvlc every VLC endpoint must answer 503, not panic on a nil player.
func TestVLCEndpointsWithoutLibVLC(t *testing.T) {
	s := &APIServer{}
	for name, h := range map[string]http.HandlerFunc{
		"load":   s.handleVLCLoad,
		"pause":  s.handleVLCPause,
		"resume": s.handleVLCResume,
		"time":   s.handleVLCTime,
		"seek":   s.handleVLCSeek,
		"stop":   s.handleVLCStop,
		"volume": s.handleVLCVolume,
	} {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodPost, "/api/v1/vlc/"+name, strings.NewReader(`{"volume":50}`)))
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: status %d, want 503", name, rec.Code)
		}
	}
}

func TestHandshakeNoiseFilterDropsOnlyEOFHandshakes(t *testing.T) {
	var buf strings.Builder
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)
	f := handshakeNoiseFilter{}
	f.Write([]byte("2026/10/13 http: TLS handshake error from 127.0.0.1:45758: EOF\n"))
	if buf.Len() != 0 {
		t.Fatalf("EOF handshake noise was logged: %q", buf.String())
	}
	f.Write([]byte("2026/10/13 http: TLS handshake error from 127.0.0.1:1: tls: bad certificate\n"))
	f.Write([]byte("2026/10/13 http: panic serving 127.0.0.1:1: boom\n"))
	if !strings.Contains(buf.String(), "bad certificate") || !strings.Contains(buf.String(), "panic serving") {
		t.Fatalf("real server errors must pass through: %q", buf.String())
	}
}
