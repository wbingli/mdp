package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	url := fmt.Sprintf("%s%s", serverURL(), filePath)
	fmt.Println(url)

	// Open in browser
	return exec.Command("open", url).Run()
}
