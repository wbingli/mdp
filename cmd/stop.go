package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running mdp server",
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	pid, err := stopServer()
	if err != nil {
		return err
	}
	if pid == 0 {
		fmt.Println("No running mdp server found.")
		return nil
	}
	fmt.Printf("mdp server stopped (pid %d)\n", pid)
	return nil
}
