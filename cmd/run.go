package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newRunCmd builds the "run" subcommand.
func newRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <name>",
		Aliases: []string{"shell", "ssh", "enter"},
		Short:   "Open an interactive shell in a running sandbox",
		Long: `Open an interactive shell session in the specified sandbox VM.
Uses the cached Debian image and SSH to the running instance.`,
		Args: cobra.ExactArgs(1),
		Run:  runRun,
	}
	runCmd.Flags().Bool("no-workspace", false, "Do not sync the current directory into /workspace")
	return runCmd
}

func runRun(cmd *cobra.Command, args []string) {
	inst := requireRunning(args[0])
	noWorkspace, _ := cmd.Flags().GetBool("no-workspace")
	cwd, _ := os.Getwd()
	if qemu.WorkspaceSyncEnabled(inst, configPath(cmd), noWorkspace) {
		if err := syncWorkspaceIn(inst, cwd); err != nil {
			printError(fmt.Sprintf("Failed to sync workspace: %v", err))
			os.Exit(1)
		}
	}

	fmt.Println(styleBanner(asciiBanner, versionString()))

	if err := qemu.SSHInteractive(inst); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to SSH: %v\n", err)
		os.Exit(1)
	}

	printSessionEnd("Session ended. Bye for now!")
}

// syncWorkspaceIn copies the current host directory into the instance workspace.
func syncWorkspaceIn(inst *qemu.Instance, cwd string) error {
	qemu.ProgressStart()
	qemu.ProgressPhase("Syncing workspace")
	err := qemu.SyncWorkspaceIn(inst, cwd)
	qemu.ProgressDone("Workspace synced to /workspace")
	qemu.ProgressStop()
	return err
}
