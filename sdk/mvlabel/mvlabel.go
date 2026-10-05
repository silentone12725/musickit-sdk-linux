// Package mvlabel holds the most recently selected MV video quality label
// (e.g. "1920x1080"). It is a shared data cell with no other dependencies
// so that sdk/apple can write it and the server can read it
// without sdk/apple needing to import engine/sdk/aacstream.
package mvlabel

import "sync"

var mu sync.RWMutex
var label string

// Set records the resolution string of the most recently selected MV variant.
func Set(resolution string) {
	mu.Lock()
	label = resolution
	mu.Unlock()
}

// Get returns the cached MV quality label, e.g. "1920x1080". Empty if unset.
func Get() string {
	mu.RLock()
	defer mu.RUnlock()
	return label
}
