package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <name>",
	Short: "Copy the sandbox workspace back to the current directory",
	Long: `Copy /workspace from the sandbox into the current directory, overwriting
matching files with the versions from the VM. This returns changes made inside
the sandbox to the host. It pairs with the automatic workspace sync that runs
on 'boite run': run copies the host directory in, sync copies the VM workspace
back out.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		inst, err := qemu.LoadInstanceState(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: instance '%s' not found. Create it first with 'boite create %s'\n", name, name)
			os.Exit(1)
		}

		if inst.PID == 0 {
			fmt.Fprintf(os.Stderr, "Error: sandbox '%s' is not running\n", name)
			os.Exit(1)
		}

		cwd, _ := os.Getwd()
		qemu.ProgressStart()
		qemu.ProgressPhase("Syncing workspace out")
		if err := qemu.SyncWorkspaceOut(inst, cwd); err != nil {
			qemu.ProgressStop()
			printError(fmt.Sprintf("Failed to sync workspace out: %v", err))
			os.Exit(1)
		}
		qemu.ProgressDone("Workspace synced back to current directory")
		qemu.ProgressStop()
		printSuccess("Done. Changes from /workspace are now in your current directory.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
