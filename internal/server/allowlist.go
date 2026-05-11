package server

import (
	"path/filepath"
	"strings"
	"sync"
)

// Allowlist is a thread-safe set of allowed file paths.
// Markdown files are always allowed (the server only binds to loopback).
// Non-markdown assets are allowed if they reside in the same directory
// (or a subdirectory) of any markdown file that has been served.
type Allowlist struct {
	mu    sync.RWMutex
	paths map[string]struct{}
}

func NewAllowlist() *Allowlist {
	return &Allowlist{paths: make(map[string]struct{})}
}

// Allow adds an absolute file path to the allowlist. Used to register
// markdown files so that assets in their directory become servable.
func (a *Allowlist) Allow(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.paths[path] = struct{}{}
}

// IsAllowed reports whether path may be served.
func (a *Allowlist) IsAllowed(path string) bool {
	if isMarkdown(path) {
		return true
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	if _, ok := a.paths[path]; ok {
		return true
	}

	// Non-markdown asset: allow if it's under any registered file's directory.
	for p := range a.paths {
		dir := filepath.Dir(p) + string(filepath.Separator)
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	return false
}
