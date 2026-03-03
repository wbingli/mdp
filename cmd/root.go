package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wbingli/mdp/internal/pidfile"
)

var (
	port int
	host string
)

var rootCmd = &cobra.Command{
	Use:   "mdp",
	Short: "Markdown preview server",
	Long:  "A local server that previews any markdown file by path with live reload.",
}

func init() {
	rootCmd.PersistentFlags().IntVar(&port, "port", 6419, "server port")
	rootCmd.PersistentFlags().StringVar(&host, "host", "localhost", "server host")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serverAddr() string {
	return fmt.Sprintf("%s:%d", host, port)
}

func serverURL() string {
	return fmt.Sprintf("http://%s:%d", host, port)
}

func mdpDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/.mdp"
	}
	return filepath.Join(home, ".mdp")
}

func pidPath() string {
	return filepath.Join(mdpDir(), "mdp.pid")
}

func isServerHealthy() bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("%s/healthz", serverURL()))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// stopServer sends SIGTERM to the running server and waits for it to exit.
// Returns the PID that was stopped, or 0 if no server was running.
func stopServer() (int, error) {
	pid, alive := pidfile.Read(pidPath())
	if !alive {
		return 0, nil
	}

	// Verify the process is actually an mdp server, not a reused PID
	if !isServerHealthy() {
		pidfile.Remove(pidPath())
		return 0, nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, fmt.Errorf("cannot find process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return 0, fmt.Errorf("cannot stop server (pid %d): %w", pid, err)
	}

	// Wait for process to exit
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			return pid, nil
		}
	}
	return pid, fmt.Errorf("server (pid %d) did not exit within 5 seconds", pid)
}
