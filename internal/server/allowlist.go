package server

import (
	"path/filepath"
	"strings"
	"sync"
)

// Allowlist is a thread-safe set of allowed file paths.
// Markdown files must match exactly; non-markdown assets are allowed
// if they reside in the same directory (or a subdirectory) of any
// allowed markdown file.
type Allowlist struct {
	mu    sync.RWMutex
	paths map[string]struct{}
}

func NewAllowlist() *Allowlist {
	return &Allowlist{paths: make(map[string]struct{})}
}

// Allow adds an absolute file path to the allowlist.
func (a *Allowlist) Allow(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.paths[path] = struct{}{}
}

// IsAllowed reports whether path may be served.
// Markdown files must be exactly on the list.
// Non-markdown files are allowed when they sit under the directory
// of any allowed markdown file.
func (a *Allowlist) IsAllowed(path string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if _, ok := a.paths[path]; ok {
		return true
	}

	if isMarkdown(path) {
		return false
	}

	// Non-markdown asset: allow if it's under any allowed file's directory.
	for p := range a.paths {
		dir := filepath.Dir(p) + string(filepath.Separator)
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	return false
}
