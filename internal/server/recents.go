package server

import (
	"sync"
	"time"
)

type RecentItem struct {
	Path      string
	Timestamp time.Time
}

type RecentsList struct {
	mu    sync.Mutex
	items []RecentItem
	cap   int
}

func NewRecentsList(cap int) *RecentsList {
	return &RecentsList{
		items: make([]RecentItem, 0, cap),
		cap:   cap,
	}
}

func (r *RecentsList) Add(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove existing entry for dedup
	for i, item := range r.items {
		if item.Path == path {
			r.items = append(r.items[:i], r.items[i+1:]...)
			break
		}
	}

	// Prepend
	entry := RecentItem{Path: path, Timestamp: time.Now()}
	r.items = append([]RecentItem{entry}, r.items...)

	// Cap
	if len(r.items) > r.cap {
		r.items = r.items[:r.cap]
	}
}

func (r *RecentsList) List() []RecentItem {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RecentItem, len(r.items))
	copy(out, r.items)
	return out
}
