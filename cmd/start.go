package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newStartCmd builds the "start" subcommand.
func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <name>",
		Short: "Start an existing sandbox",
		Long:  `Start a previously created sandbox VM.`,
		Args:  cobra.ExactArgs(1),
		Run:   runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) {
	name := args[0]
	qemu.ProgressStart()
	qemu.ProgressPhase("Starting VM")
	_, err := qemu.Start(name, configPath(cmd))
	qemu.ProgressStop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to start sandbox '%s': %v\n", name, err)
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("Sandbox '%s' started", name))
	printInfo(fmt.Sprintf("Connect: boite run %s", name))
}
