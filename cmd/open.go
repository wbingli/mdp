package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
