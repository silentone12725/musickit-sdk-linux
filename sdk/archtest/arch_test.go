// Package archtest verifies the SDK's import-boundary invariants.
//
// They live in a subpackage of the sdk module and shell out to `go list`.
//
// All test names are prefixed TestArch so CI can select them with
//
//	go test -run TestArch ./...
package archtest

import (
	"os/exec"
	"strings"
	"testing"
)

const repoRoot = ".."

// directImports returns the packages that pkg imports directly (not transitively).
func directImports(t *testing.T, pkg string) []string {
	t.Helper()
	return goList(t, "-f", "{{range .Imports}}{{.}}\n{{end}}", pkg)
}

// deps returns the full transitive dependency list of pkg.
func deps(t *testing.T, pkg string) []string {
	t.Helper()
	return goList(t, "-deps", pkg)
}

func goList(t *testing.T, args ...string) []string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"list"}, args...)...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list %v: %v\n%s", args, err, out)
	}
	var pkgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			pkgs = append(pkgs, s)
		}
	}
	return pkgs
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// utilPackages are the helper packages (formerly engine/utils/*) plus test and
// internal-only packages. The boundary rules below apply to the remaining
// "core" packages; helpers are allowed their documented bridges.
var utilPackages = map[string]bool{
	"aacstream": true, "alacstream": true, "ampapi": true, "config": true,
	"lyrics": true, "mvlabel": true, "archtest": true, "internal": true, "ring": true,
}

// enginePackages lists the sdk's core packages (every package except the helper
// set above) via `go list`.
func enginePackages(t *testing.T) []string {
	t.Helper()
	const prefix = "github.com/silentone12725/musickit-sdk-linux/sdk/"
	all := goList(t, "./...")
	// Guard: a stale pattern would silently turn every rule below into a no-op.
	if len(all) < 10 {
		t.Fatalf("go list ./... returned only %d packages (%v); the package pattern is out of date", len(all), all)
	}
	var core []string
	for _, pkg := range all {
		rel := strings.TrimPrefix(pkg, prefix)
		if utilPackages[strings.SplitN(rel, "/", 2)[0]] {
			continue
		}
		core = append(core, pkg)
	}
	if len(core) < 8 {
		t.Fatalf("only %d core packages found (%v); the helper-package list is out of date", len(core), core)
	}
	return core
}

// TestArchRunv3ImportedOnlyByFairplay: among engine packages, only
// sdk/fairplay may import sdk/aacstream directly.
func TestArchRunv3ImportedOnlyByFairplay(t *testing.T) {
	const runv3 = "github.com/silentone12725/musickit-sdk-linux/sdk/aacstream"
	for _, pkg := range enginePackages(t) {
		imports := directImports(t, pkg)
		if contains(imports, runv3) && pkg != "github.com/silentone12725/musickit-sdk-linux/sdk/fairplay" {
			t.Errorf("%s directly imports %s; only sdk/fairplay may", pkg, runv3)
		}
	}
	// And fairplay must actually import it (guards against silent refactors).
	if !contains(directImports(t, "github.com/silentone12725/musickit-sdk-linux/sdk/fairplay"), runv3) {
		t.Errorf("sdk/fairplay no longer imports %s", runv3)
	}
}

func TestArchPlaybackDoesNotImportRunv3(t *testing.T) {
	if contains(directImports(t, "github.com/silentone12725/musickit-sdk-linux/sdk/playback"), "github.com/silentone12725/musickit-sdk-linux/sdk/aacstream") {
		t.Error("sdk/playback must not directly import sdk/aacstream")
	}
}

func TestArchAppleDoesNotImportRunv3(t *testing.T) {
	if contains(directImports(t, "github.com/silentone12725/musickit-sdk-linux/sdk/apple"), "github.com/silentone12725/musickit-sdk-linux/sdk/aacstream") {
		t.Error("sdk/apple must not directly import sdk/aacstream")
	}
}

func TestArchHLSDoesNotImportRunv3(t *testing.T) {
	if contains(directImports(t, "github.com/silentone12725/musickit-sdk-linux/sdk/hls"), "github.com/silentone12725/musickit-sdk-linux/sdk/aacstream") {
		t.Error("sdk/hls must not directly import sdk/aacstream")
	}
}

// TestArchMediaHasNoAppleSpecificDeps: sdk/media is the provider-agnostic
// abstraction; none of its transitive deps may be Apple/DRM specific.
func TestArchMediaHasNoAppleSpecificDeps(t *testing.T) {
	forbidden := []string{"apple", "ampapi", "aacstream", "fairplay"}
	for _, dep := range deps(t, "github.com/silentone12725/musickit-sdk-linux/sdk/media") {
		for _, bad := range forbidden {
			if strings.Contains(dep, bad) {
				t.Errorf("sdk/media depends on %q (contains %q)", dep, bad)
			}
		}
	}
}

// TestArchNoCycles: `go build ./...` succeeds (import cycles fail the build).
func TestArchNoCycles(t *testing.T) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./... failed (possible import cycle):\n%s", out)
	}
}

// TestArchEngineDoesNotImportCmd: sdk packages must never import the server module
// (CLI tools). A front end depends on the sdk, never the reverse.
func TestArchEngineDoesNotImportCmd(t *testing.T) {
	for _, pkg := range enginePackages(t) {
		for _, dep := range deps(t, pkg) {
			if strings.HasPrefix(dep, "github.com/silentone12725/musickit-sdk-linux/server/") {
				t.Errorf("%s depends on %s (the server module must not be imported by the sdk)", pkg, dep)
			}
		}
	}
}

// TestArchDRMDoesNotImportFairplay: sdk/drm must not import sdk/fairplay.
// CBCSDialer is defined in sdk/fairplay so that fairplay can define it without
// importing drm — enforcing a one-way drm → ??? dependency (drm is standalone).
func TestArchDRMDoesNotImportFairplay(t *testing.T) {
	if contains(deps(t, "github.com/silentone12725/musickit-sdk-linux/sdk/drm"), "github.com/silentone12725/musickit-sdk-linux/sdk/fairplay") {
		t.Error("sdk/drm must not import sdk/fairplay (CBCSDialer lives in fairplay to prevent this cycle)")
	}
}

// TestArchFairplayDoesNotImportDRM: sdk/fairplay must not import sdk/drm.
// CBCSDialer is defined in sdk/fairplay so that fairplay remains importable
// without pulling in the full DRM subsystem.
func TestArchFairplayDoesNotImportDRM(t *testing.T) {
	if contains(deps(t, "github.com/silentone12725/musickit-sdk-linux/sdk/fairplay"), "github.com/silentone12725/musickit-sdk-linux/sdk/drm") {
		t.Error("sdk/fairplay must not import sdk/drm")
	}
}

// TestArchEngineDoesNotImportTUI: engine packages must never import the TUI
// layer (tui.go, survey, tablewriter pulled in only by the CLI).
func TestArchEngineDoesNotImportTUI(t *testing.T) {
	forbidden := []string{
		"github.com/AlecAivazis/survey",
		"github.com/olekukonko/tablewriter",
		"github.com/fatih/color",
	}
	for _, pkg := range enginePackages(t) {
		for _, dep := range deps(t, pkg) {
			for _, bad := range forbidden {
				if strings.HasPrefix(dep, bad) {
					t.Errorf("%s depends on %q (TUI-only dependency)", pkg, dep)
				}
			}
		}
	}
}

// TestArchEngineDoesNotDirectlyImportConfig: engine packages must not
// DIRECTLY import sdk/config (the CLI ConfigSet).  Transitive exposure
// through sdk/alacstream is a known Temporary Bridge (see docs/dependency-audit.md
// §Phase3); it will be eliminated when alacstream functions are moved into
// sdk/fairplay.  Until then only direct imports are enforced here.
func TestArchEngineDoesNotDirectlyImportStructs(t *testing.T) {
	const config = "github.com/silentone12725/musickit-sdk-linux/sdk/config"
	for _, pkg := range enginePackages(t) {
		if contains(directImports(t, pkg), config) {
			t.Errorf("%s directly imports %s (CLI config must not be a direct engine dep)", pkg, config)
		}
	}
	// Document the known transitive path so it's not invisible:
	// sdk/fairplay → sdk/alacstream → sdk/config.
	// Upgrade this check to use deps() once Phase 3 is complete.
}

// TestArchAlacstreamImportedOnlyByFairplay: only sdk/fairplay may import
// sdk/alacstream directly. This matches the dependency-audit ISL/TB classification.
// When Phase 3 of the decommissioning plan completes, this test should be
// updated to assert that sdk/fairplay no longer imports alacstream either.
func TestArchRunv2ImportedOnlyByFairplay(t *testing.T) {
	const alacstream = "github.com/silentone12725/musickit-sdk-linux/sdk/alacstream"
	for _, pkg := range enginePackages(t) {
		if contains(directImports(t, pkg), alacstream) && pkg != "github.com/silentone12725/musickit-sdk-linux/sdk/fairplay" {
			t.Errorf("%s directly imports %s; only sdk/fairplay may", pkg, alacstream)
		}
	}
	// Guard: fairplay must actually import it until Phase 3 is complete.
	if !contains(directImports(t, "github.com/silentone12725/musickit-sdk-linux/sdk/fairplay"), alacstream) {
		t.Log("sdk/fairplay no longer imports alacstream — Phase 3 may be complete; update this test")
	}
}

// TestArchExportDoesNotDirectlyImportEngineInternals: sdk/export interacts
// with media acquisition only through the sdk/playback.Manager API.  It
// must not directly import sdk/apple, sdk/hls, or sdk/fairplay —
// those are playback-internal packages.
// Note: transitive exposure through sdk/playback is expected and permitted.
func TestArchExportDoesNotDirectlyImportEngineInternals(t *testing.T) {
	exportImports := directImports(t, "github.com/silentone12725/musickit-sdk-linux/sdk/export")
	forbidden := []string{"github.com/silentone12725/musickit-sdk-linux/sdk/apple", "github.com/silentone12725/musickit-sdk-linux/sdk/hls", "github.com/silentone12725/musickit-sdk-linux/sdk/fairplay"}
	for _, imp := range exportImports {
		for _, bad := range forbidden {
			if imp == bad {
				t.Errorf("sdk/export directly imports %q — use sdk/playback.Manager API instead", bad)
			}
		}
	}
}
