package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirsUseTheSDKName(t *testing.T) {
	config, cache := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("XDG_CACHE_HOME", cache)
	resolved = map[bool]string{}

	if got, want := Config(), filepath.Join(config, Name); got != want {
		t.Errorf("Config() = %q, want %q", got, want)
	}
	if got, want := Cache(), filepath.Join(cache, Name); got != want {
		t.Errorf("Cache() = %q, want %q", got, want)
	}
}

func TestMoveIntoRelocatesOnlyWhenTargetIsFree(t *testing.T) {
	dir := t.TempDir()
	old, dst := filepath.Join(dir, "old"), filepath.Join(dir, "sub", "new")
	if err := writeFile(old, "a"); err != nil {
		t.Fatal(err)
	}
	MoveInto(old, dst)
	if b, err := readFile(dst); err != nil || b != "a" {
		t.Fatalf("target = %q, %v; want the moved file", b, err)
	}
	if err := writeFile(old, "b"); err != nil {
		t.Fatal(err)
	}
	MoveInto(old, dst)
	if b, _ := readFile(dst); b != "a" {
		t.Fatalf("an existing target was overwritten: %q", b)
	}
}

func writeFile(p, s string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(s), 0o600)
}

func readFile(p string) (string, error) {
	b, err := os.ReadFile(p)
	return string(b), err
}
