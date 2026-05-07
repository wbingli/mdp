package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/wbingli/mdp/internal/watcher"
)

type Server struct {
	Addr      string
	Recents   *RecentsList
	Watcher   *watcher.Hub
	Allowlist *Allowlist
	srv       *http.Server
}

func New(addr string) *Server {
	s := &Server{
		Addr:      addr,
		Recents:   NewRecentsList(50),
		Watcher:   watcher.NewHub(),
		Allowlist: NewAllowlist(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/allow", s.handleAllow)
	mux.HandleFunc("/events/", s.handleSSE)
	mux.HandleFunc("/", s.handleCatchAll)

	s.srv = &http.Server{
		Addr:        addr,
		Handler:     mux,
		IdleTimeout: 60 * time.Second,
	}
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.Addr, err)
	}
	log.Printf("Server listening on %s", s.Addr)
	return s.srv.Serve(ln)
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Watcher.Close()
	return s.srv.Shutdown(ctx)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func (s *Server) handleAllow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	filePath := strings.TrimSpace(string(body))
	if filePath == "" {
		http.Error(w, "empty path", http.StatusBadRequest)
		return
	}
	s.Allowlist.Allow(filePath)
	log.Printf("Allowlisted: %s", filePath)
	w.WriteHeader(http.StatusOK)
}
