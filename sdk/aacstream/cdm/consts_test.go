package wv

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIdentity(t *testing.T) {
	dir := t.TempDir()
	if err := loadIdentity(dir); !errors.Is(err, ErrNoDeviceIdentity) {
		t.Fatalf("empty dir: want ErrNoDeviceIdentity, got %v", err)
	}
	os.WriteFile(filepath.Join(dir, "device_private_key"), []byte("KEY"), 0o600)
	if err := loadIdentity(dir); !errors.Is(err, ErrNoDeviceIdentity) {
		t.Fatalf("missing client id: want ErrNoDeviceIdentity, got %v", err)
	}
	os.WriteFile(filepath.Join(dir, "device_client_id_blob"), []byte{1, 2}, 0o600)
	if err := loadIdentity(dir); err != nil {
		t.Fatal(err)
	}
	if DefaultPrivateKey != "KEY" || len(DefaultClientID) != 2 {
		t.Fatal("identity not loaded")
	}
}
