package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newSyncCmd builds the "sync" subcommand.
func newSyncCmd() *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync <name>",
		Short: "Copy the sandbox workspace back to the current directory",
		Long: `Copy /workspace from the sandbox into the current directory, overwriting
matching files with the versions from the VM. This returns changes made inside
the sandbox to the host. It pairs with the automatic workspace sync that runs
on 'boite run': run copies the host directory in, sync copies the VM workspace
back out.

A deny list keeps guest-authored hooks and env files (.git/hooks/*, .envrc,
.env, .env.*, .vscode/tasks.json, .husky/*) and setuid, setgid or
world-writable files out of the host tree; every skip prints a warning line.
Pass --all to disable the filter for this one sync and copy everything.`,
		Args: cobra.ExactArgs(1),
		Run:  runSync,
	}
	syncCmd.Flags().Bool("all", false, "Copy everything, bypassing the sync-out deny list (guest hooks, env files, setuid or world-writable files)")
	return syncCmd
}

func runSync(cmd *cobra.Command, args []string) {
	name := args[0]
	inst, err := qemu.LoadInstanceState(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: instance '%s' not found. Create it first with 'boite create %s'\n", name, name)
		os.Exit(1)
	}

	if inst.PID == 0 || !qemu.IsProcessRunning(inst.PID) {
		fmt.Fprintf(os.Stderr, "Error: sandbox '%s' is not running\n", name)
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	all, _ := cmd.Flags().GetBool("all")
	qemu.ProgressStart()
	qemu.ProgressPhase("Syncing workspace out")
	if err := qemu.SyncWorkspaceOut(inst, cwd, all); err != nil {
		qemu.ProgressStop()
		printError(fmt.Sprintf("Failed to sync workspace out: %v", err))
		os.Exit(1)
	}
	qemu.ProgressDone("Workspace synced back to current directory")
	qemu.ProgressStop()
	printSuccess("Done. Changes from /workspace are now in your current directory.")
}
