package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/wbingli/mdp/internal/watcher"
)

type Server struct {
	Addr    string
	Recents *RecentsList
	Watcher *watcher.Hub
	srv     *http.Server
}

func New(addr string) *Server {
	s := &Server{
		Addr:    addr,
		Recents: NewRecentsList(50),
		Watcher: watcher.NewHub(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/events/", s.handleSSE)
	mux.HandleFunc("/", s.handleCatchAll)

	s.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
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
