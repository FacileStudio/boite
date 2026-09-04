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
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n")
			os.Exit(1)
		}

		if err := runPurge(); err != nil {
			fmt.Fprintf(os.Stderr, "Error purging deleted VMs: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All deleted VMs have been purged")
	},
}

func runPurge() error {
	_, err := RunCommand("multipass", "purge")
	return err
}

func init() {
	rootCmd.AddCommand(purgeCmd)
}
