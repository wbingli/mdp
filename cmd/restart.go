package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wbingli/mdp/internal/pidfile"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the mdp server",
	RunE:  runRestart,
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
	pid, err := stopServer()
	if err != nil {
		return err
	}
	if pid > 0 {
		fmt.Printf("Stopped mdp server (pid %d)\n", pid)
	}

	if err := startBackgroundAndWait(); err != nil {
		return err
	}

	newPid, _ := pidfile.Read(pidPath())
	fmt.Printf("mdp server started on %s (pid %d)\n", serverURL(), newPid)
	return nil
}
