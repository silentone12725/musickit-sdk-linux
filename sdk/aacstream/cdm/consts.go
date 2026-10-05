package wv

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// The Widevine device identity is supplied by the embedder; the SDK ships none.
//
// Two files are read from the identity directory:
//
//	device_private_key       PEM RSA private key
//	device_client_id_blob    serialised ClientIdentification protobuf (raw bytes)
//
// The directory is $MUSICKIT_WIDEVINE_DIR, or <user config dir>/musickit-sdk-linux/widevine.

// ErrNoDeviceIdentity is returned when no device identity could be loaded.
var ErrNoDeviceIdentity = errors.New("widevine device identity not configured")

var (
	DefaultPrivateKey string
	DefaultClientID   []byte

	identityErr error
	identityMu  sync.Mutex
)

// IdentityDir returns the directory the device identity is loaded from.
func IdentityDir() string {
	if d := os.Getenv("MUSICKIT_WIDEVINE_DIR"); d != "" {
		return d
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "musickit-sdk-linux", "widevine")
}

// InitConstants loads the device identity. It is safe to call repeatedly; a
// failure is reported by NewDefaultCDM as an error wrapping ErrNoDeviceIdentity.
func InitConstants() {
	identityMu.Lock()
	defer identityMu.Unlock()
	identityErr = loadIdentity(IdentityDir())
}

func loadIdentity(dir string) error {
	if dir == "" {
		return fmt.Errorf("%w: no config directory", ErrNoDeviceIdentity)
	}
	key, err := os.ReadFile(filepath.Join(dir, "device_private_key"))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoDeviceIdentity, err)
	}
	cid, err := os.ReadFile(filepath.Join(dir, "device_client_id_blob"))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoDeviceIdentity, err)
	}
	DefaultPrivateKey, DefaultClientID = string(key), cid
	return nil
}
