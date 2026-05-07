package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wbingli/mdp/internal/render"
)

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// /events/<filepath> → extract filepath
	filePath := strings.TrimPrefix(r.URL.Path, "/events")
	if filePath == "" || filePath == "/" {
		http.Error(w, "file path required", http.StatusBadRequest)
		return
	}

	if !s.Allowlist.IsAllowed(filePath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher.Flush()

	ch := s.Watcher.Subscribe(filePath)
	defer s.Watcher.Unsubscribe(filePath, ch)

	// Heartbeat ensures the server detects client disconnection promptly.
	// Without periodic writes, a disconnected client may not be detected
	// until the next file-change event, causing the connection to linger
	// and potentially blocking new requests on page refresh.
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// SSE comment line — ignored by the browser's EventSource API
			if _, err := fmt.Fprintf(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case _, ok := <-ch:
			if !ok {
				return
			}
			source, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("SSE: cannot read %s: %v", filePath, err)
				continue
			}
			htmlContent, err := render.ToHTML(source)
			if err != nil {
				log.Printf("SSE: render error for %s: %v", filePath, err)
				continue
			}
			// SSE multi-line data: each line prefixed with "data:"
			lines := strings.Split(string(htmlContent), "\n")
			fmt.Fprintf(w, "event: update\n")
			for _, line := range lines {
				fmt.Fprintf(w, "data: %s\n", line)
			}
			fmt.Fprintf(w, "\n")
			flusher.Flush()
		}
	}
}
