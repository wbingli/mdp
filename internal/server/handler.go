package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wbingli/mdp/internal/render"
)

func isMarkdown(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown" || ext == ".mdown" || ext == ".mkd"
}

func (s *Server) handleCatchAll(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	// Homepage
	if urlPath == "/" {
		s.handleIndex(w, r)
		return
	}

	// Treat the URL path as an absolute file path
	filePath := urlPath

	info, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("File not found: %s", filePath), http.StatusNotFound)
		return
	}
	if info.IsDir() {
		// Try index.md in the directory
		indexPath := filepath.Join(filePath, "README.md")
		if _, err := os.Stat(indexPath); err == nil {
			filePath = indexPath
		} else {
			http.Error(w, "Directory listing not supported", http.StatusForbidden)
			return
		}
	}

	if !s.Allowlist.IsAllowed(filePath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !isMarkdown(filePath) {
		http.ServeFile(w, r, filePath)
		return
	}

	s.renderMarkdown(w, r, filePath)
}

func (s *Server) renderMarkdown(w http.ResponseWriter, r *http.Request, filePath string) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot read file: %v", err), http.StatusInternalServerError)
		return
	}

	htmlContent, err := render.ToHTML(source)
	if err != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	s.Recents.Add(filePath)

	tmpl, err := template.New("preview.html").Parse(string(previewHTML))
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	fileName := filepath.Base(filePath)
	data := struct {
		Title    string
		FilePath string
		Content  template.HTML
	}{
		Title:    fileName,
		FilePath: filePath,
		Content:  template.HTML(htmlContent),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execute error: %v", err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index.html").Parse(string(indexHTML))
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Recents []RecentItem
	}{
		Recents: s.Recents.List(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execute error: %v", err)
	}
}
