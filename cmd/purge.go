package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge all deleted development sandbox VMs",
	Long: `Purge completely removes all deleted sandbox VMs, freeing disk space.
This action cannot be undone. It purges every VM that is in the 'Deleted' state.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkCommand("multipass"); err != nil {
			printError("multipass not found in PATH")
			os.Exit(1)
		}

		if err := runPurge(); err != nil {
			printError(fmt.Sprintf("Failed to purge deleted sandboxes: %v", err))
			os.Exit(1)
		}
		printSuccess("All deleted sandbox VMs have been purged")
	},
}

func runPurge() error {
	_, err := runCommand("multipass", "purge")
	return err
}

func init() {
	rootCmd.AddCommand(purgeCmd)
}
