package cmd

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newListCmd builds the "list" subcommand.
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all development sandboxes",
		Long:  `List all development sandboxes managed by boite.`,
		Run:   runList,
	}
}

func runList(cmd *cobra.Command, args []string) {
	instances, err := qemu.ListInstances()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list sandboxes: %v\n", err)
		os.Exit(1)
	}
	lipgloss.Print(renderInstanceTableQEMU(instances))
}
