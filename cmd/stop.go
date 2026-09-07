package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running sandbox",
	Long:  `Stop a running sandbox VM (keeps the overlay disk).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := qemu.Stop(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to stop sandbox '%s': %v\n", name, err)
			os.Exit(1)
		}
		printSuccess(fmt.Sprintf("Sandbox '%s' stopped", name))
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}