package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newExecCmd builds the "exec" subcommand.
func newExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec <name> <command...>",
		Short: "Execute a command in a running sandbox",
		Long:  `Execute a command in the specified sandbox VM via SSH.`,
		Args:  cobra.MinimumNArgs(1),
		Run:   runExec,
	}
}

func runExec(cmd *cobra.Command, args []string) {
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

	if err := RefreshManagedEnv(name, configPath(cmd)); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not refresh managed env, keeping last snapshot: %v\n", err)
	}

	command := args[1:]
	if err := qemu.SSHCommand(inst, qemu.WithEnv(command)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: command execution failed: %v\n", err)
		os.Exit(1)
	}
}
