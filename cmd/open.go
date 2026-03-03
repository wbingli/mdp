package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <file>",
	Short: "Preview a markdown file in the browser",
	Args:  cobra.ExactArgs(1),
	RunE:  runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	filePath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Ensure server is running
	if !isServerHealthy() {
		fmt.Fprintln(os.Stderr, "Starting mdp server...")
		if err := startBackgroundAndWait(); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	}

	// Register the file on the server's allowlist
	allowURL := fmt.Sprintf("%s/api/allow", serverURL())
	resp, err := http.Post(allowURL, "text/plain", strings.NewReader(filePath))
	if err != nil {
		return fmt.Errorf("failed to register file: %w", err)
	}
	resp.Body.Close()

	u := &url.URL{
		Scheme: "http",
		Host:   serverAddr(),
		Path:   filePath,
	}
	fmt.Println(u.String())

	// Open in browser (macOS only)
	if runtime.GOOS == "darwin" {
		return exec.Command("open", u.String()).Run()
	}
	return nil
}
