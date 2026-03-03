package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wbingli/mdp/internal/pidfile"
	"github.com/wbingli/mdp/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the markdown preview server in the foreground",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	return startForeground()
}

func startForeground() error {
	addr := serverAddr()

	// Set up logging
	logDir := mdpDir()
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "server.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}

	// Write PID file
	pidPath := filepath.Join(logDir, "mdp.pid")
	if err := pidfile.Write(pidPath); err != nil {
		log.Printf("Warning: could not write PID file: %v", err)
	}
	defer pidfile.Remove(pidPath)

	s := server.New(addr)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		s.Shutdown()
	}()

	fmt.Fprintf(os.Stderr, "mdp server listening on %s\n", serverURL())
	return s.Start()
}

// startBackgroundAndWait launches the server in the background and polls until healthy.
func startBackgroundAndWait() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	logDir := mdpDir()
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "server.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("cannot open log file: %w", err)
	}

	attr := &os.ProcAttr{
		Dir:   "/",
		Files: []*os.File{os.Stdin, logFile, logFile},
		Sys: &syscall.SysProcAttr{
			Setpgid: true,
		},
	}

	proc, err := os.StartProcess(exe, []string{exe, "serve", "--port", fmt.Sprintf("%d", port), "--host", host}, attr)
	if err != nil {
		logFile.Close()
		return fmt.Errorf("cannot start background process: %w", err)
	}

	logFile.Close()
	proc.Release()

	// Poll until healthy
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if isServerHealthy() {
			return nil
		}
	}
	return fmt.Errorf("server did not become healthy within 3 seconds")
}

