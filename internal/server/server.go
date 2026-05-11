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
	if err := requireLoopback(s.Addr); err != nil {
		return err
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.Addr, err)
	}
	log.Printf("Server listening on %s", s.Addr)
	return s.srv.Serve(ln)
}

// requireLoopback returns an error unless every address that addr's host
// resolves to is a loopback address. mdp serves arbitrary local files, so
// it must never bind to an externally reachable interface.
func requireLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid addr %q: %w", addr, err)
	}
	if host == "" {
		return fmt.Errorf("host must be a loopback address; empty host binds to all interfaces")
	}
	addrs, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("cannot resolve host %q: %w", host, err)
	}
	for _, a := range addrs {
		ip := net.ParseIP(a)
		if ip == nil || !ip.IsLoopback() {
			return fmt.Errorf("host %q resolves to non-loopback address %s; mdp only binds to localhost", host, a)
		}
	}
	return nil
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
