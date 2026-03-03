package watcher

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Hub manages per-directory file watchers and fans out change notifications
// to subscribers keyed by file path.
type Hub struct {
	mu          sync.Mutex
	watcher     *fsnotify.Watcher
	subscribers map[string][]chan struct{} // filepath -> subscriber channels
	watchedDirs map[string]bool
	closed      bool
}

func NewHub() *Hub {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Warning: could not create file watcher: %v", err)
		return &Hub{
			subscribers: make(map[string][]chan struct{}),
			watchedDirs: make(map[string]bool),
		}
	}

	h := &Hub{
		watcher:     w,
		subscribers: make(map[string][]chan struct{}),
		watchedDirs: make(map[string]bool),
	}
	go h.loop()
	return h
}

func (h *Hub) loop() {
	if h.watcher == nil {
		return
	}

	debounce := make(map[string]*time.Timer)

	for {
		select {
		case event, ok := <-h.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				filePath := event.Name
				if t, exists := debounce[filePath]; exists {
					t.Stop()
				}
				debounce[filePath] = time.AfterFunc(100*time.Millisecond, func() {
					h.notify(filePath)
				})
			}
		case err, ok := <-h.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (h *Hub) notify(filePath string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, ch := range h.subscribers[filePath] {
		select {
		case ch <- struct{}{}:
		default:
			// Subscriber not ready, skip
		}
	}
}

// Subscribe creates a channel that receives notifications when the given file changes.
// The caller must call Unsubscribe when done.
func (h *Hub) Subscribe(filePath string) chan struct{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan struct{}, 1)
	h.subscribers[filePath] = append(h.subscribers[filePath], ch)

	// Watch the directory if not already watching
	dir := filepath.Dir(filePath)
	if !h.watchedDirs[dir] && h.watcher != nil {
		if err := h.watcher.Add(dir); err != nil {
			log.Printf("Warning: could not watch directory %s: %v", dir, err)
		} else {
			h.watchedDirs[dir] = true
		}
	}

	return ch
}

// Unsubscribe removes a subscriber channel for the given file path.
func (h *Hub) Unsubscribe(filePath string, ch chan struct{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs := h.subscribers[filePath]
	for i, sub := range subs {
		if sub == ch {
			h.subscribers[filePath] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	close(ch)
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	if h.watcher != nil {
		h.watcher.Close()
	}
}
