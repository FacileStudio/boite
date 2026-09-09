package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"destroy", "delete"},
	Short:   "Destroy a sandbox (delete overlay and state)",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := qemu.Destroy(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to destroy '%s': %v\n", name, err)
			os.Exit(1)
		}
		printSuccess(fmt.Sprintf("Sandbox '%s' destroyed (overlay removed)", name))
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
