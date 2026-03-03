package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wbingli/mdp/internal/pidfile"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the mdp server in the background",
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	if _, alive := pidfile.Read(pidPath()); alive {
		fmt.Printf("mdp server is already running on %s\n", serverURL())
		return nil
	}

	if err := startBackgroundAndWait(); err != nil {
		return err
	}

	pid, _ := pidfile.Read(pidPath())
	fmt.Printf("mdp server started on %s (pid %d)\n", serverURL(), pid)
	return nil
}
