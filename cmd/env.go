package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newEnvCmd builds the "env" subcommand and its children.
func newEnvCmd() *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage a sandbox's environment store",
		Long: `Manage the tiroir environment store of a running sandbox. Values written
via 'boite env set' go straight into the guest store and stick: they are
authoritative over source refresh (casier/host links are not re-applied on top).`,
	}
	envCmd.AddCommand(
		&cobra.Command{Use: "list <name>", Short: "List keys in a sandbox's environment store", Args: cobra.ExactArgs(1), Run: runEnvList},
		&cobra.Command{Use: "get <name> <key>", Short: "Print a value from a sandbox's environment store", Args: cobra.ExactArgs(2), Run: runEnvGet},
		&cobra.Command{Use: "set <name> <key> <value>", Short: "Write a value into a sandbox's environment store", Args: cobra.ExactArgs(3), Run: runEnvSet},
		&cobra.Command{Use: "delete <name> <key>", Short: "Remove a key from a sandbox's environment store", Args: cobra.ExactArgs(2), Run: runEnvDelete},
	)
	return envCmd
}

func runEnvList(cmd *cobra.Command, args []string) {
	inst := requireRunning(args[0])
	out, err := qemu.SSHOutput(inst, "tiroir list")
	if err != nil {
		printError(fmt.Sprintf("Error: failed to list env: %v", err))
		os.Exit(1)
	}
	fmt.Print(out)
}

func runEnvGet(cmd *cobra.Command, args []string) {
	inst := requireRunning(args[0])
	out, err := qemu.SSHOutput(inst, "tiroir get "+args[1])
	if err != nil {
		printError(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
	fmt.Print(out)
}

func runEnvSet(cmd *cobra.Command, args []string) {
	inst := requireRunning(args[0])
	remote := "tiroir set " + args[1] + " " + shellQuoteArg(args[2])
	if err := qemu.SSHCommand(inst, []string{remote}); err != nil {
		printError(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
	if err := pinEnv(inst, args[1]); err != nil {
		printError(fmt.Sprintf("Error: could not pin %s: %v", args[1], err))
		os.Exit(1)
	}
}

func runEnvDelete(cmd *cobra.Command, args []string) {
	inst := requireRunning(args[0])
	if err := qemu.SSHCommand(inst, []string{"tiroir delete " + args[1]}); err != nil {
		printError(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
	if err := unpinEnv(inst, args[1]); err != nil {
		printError(fmt.Sprintf("Error: could not unpin %s: %v", args[1], err))
		os.Exit(1)
	}
}

// requireRunning loads a running instance or exits with an error.
func requireRunning(name string) *qemu.Instance {
	inst, err := qemu.LoadInstanceState(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: instance '%s' not found. Create it first with 'boite create %s'\n", name, name)
		os.Exit(1)
	}
	if inst.PID == 0 {
		fmt.Fprintf(os.Stderr, "Error: sandbox '%s' is not running\n", name)
		os.Exit(1)
	}
	return inst
}

// shellQuoteArg wraps s in single quotes for passing through SSH.
func shellQuoteArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
