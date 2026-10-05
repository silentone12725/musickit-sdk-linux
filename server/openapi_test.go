package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// routeRe matches mux registrations such as mux.HandleFunc("GET /api/v1/status", ...).
var routeRe = regexp.MustCompile(`mux\.HandleFunc\("([A-Z]+) ([^"]+)"`)

// TestOpenAPICoversEveryRoute fails when a route is registered without being
// documented in api/openapi.json (or documented but no longer registered), so the
// machine-readable spec cannot drift from the server.
func TestOpenAPICoversEveryRoute(t *testing.T) {
	registered := map[string]bool{}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range routeRe.FindAllStringSubmatch(string(src), -1) {
			registered[m[1]+" "+m[2]] = true
		}
	}
	if len(registered) == 0 {
		t.Fatal("found no registered routes; the route pattern is out of date")
	}

	raw, err := os.ReadFile("../api/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("openapi.json: %v", err)
	}
	documented := map[string]bool{}
	for path, ops := range spec.Paths {
		for method := range ops {
			switch method {
			case "get", "post", "put", "delete", "patch":
				documented[strings.ToUpper(method)+" "+path] = true
			}
		}
	}

	var undocumented, stale []string
	for r := range registered {
		if !documented[r] {
			undocumented = append(undocumented, r)
		}
	}
	for d := range documented {
		if !registered[d] {
			stale = append(stale, d)
		}
	}
	sort.Strings(undocumented)
	sort.Strings(stale)
	if len(undocumented) > 0 {
		t.Errorf("routes missing from api/openapi.json:\n  %s", strings.Join(undocumented, "\n  "))
	}
	if len(stale) > 0 {
		t.Errorf("documented routes that are no longer registered:\n  %s", strings.Join(stale, "\n  "))
	}
}
