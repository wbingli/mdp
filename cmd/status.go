package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wbingli/mdp/internal/pidfile"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the mdp server",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	pid, alive := pidfile.Read(pidPath())

	if !alive {
		fmt.Println("mdp server is not running")
		return nil
	}

	healthy := isServerHealthy()

	if healthy {
		fmt.Printf("mdp server is running on %s (pid %d)\n", serverURL(), pid)
	} else {
		fmt.Printf("mdp server process exists (pid %d) but is not responding on %s\n", pid, serverURL())
	}

	return nil
}
