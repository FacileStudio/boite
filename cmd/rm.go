package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newRmCmd builds the "rm" subcommand.
func newRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <name>",
		Aliases: []string{"destroy", "delete"},
		Short:   "Destroy a sandbox (delete overlay and state)",
		Args:    cobra.ExactArgs(1),
		Run:     runRm,
	}
}

func runRm(cmd *cobra.Command, args []string) {
	name := args[0]
	qemu.ProgressStart()
	qemu.ProgressPhase("Destroying sandbox")
	err := qemu.Destroy(name)
	qemu.ProgressStop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to destroy '%s': %v\n", name, err)
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("Sandbox '%s' destroyed (overlay removed)", name))
}
