package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/wbingli/mdp/internal/render"
)

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// /events/<filepath> → extract filepath
	filePath := strings.TrimPrefix(r.URL.Path, "/events")
	if filePath == "" || filePath == "/" {
		http.Error(w, "file path required", http.StatusBadRequest)
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

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
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
