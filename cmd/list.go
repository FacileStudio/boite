package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all development sandboxes",
	Long:  `List all development sandboxes managed by boite.`,
	Run: func(cmd *cobra.Command, args []string) {
		instances, err := qemu.ListInstances()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to list sandboxes: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(renderInstanceTableQEMU(instances))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}