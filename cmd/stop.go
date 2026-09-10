package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newStopCmd builds the "stop" subcommand.
func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <name>",
		Short: "Stop a running sandbox",
		Long:  `Stop a running sandbox VM (keeps the overlay disk).`,
		Args:  cobra.ExactArgs(1),
		Run:   runStop,
	}
}

func runStop(cmd *cobra.Command, args []string) {
	name := args[0]
	qemu.ProgressStart()
	qemu.ProgressPhase("Stopping VM")
	err := qemu.Stop(name)
	qemu.ProgressStop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to stop sandbox '%s': %v\n", name, err)
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("Sandbox '%s' stopped", name))
}
