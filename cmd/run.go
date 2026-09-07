package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:     "run <name>",
	Aliases: []string{"shell", "ssh", "enter"},
	Short:   "Open an interactive shell in a running sandbox",
	Long: `Open an interactive shell session in the specified sandbox VM.
Uses the cached Debian image and SSH to the running instance.`,
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

		fmt.Println(styleBanner(asciiBanner, versionString()))

		if err := qemu.SSHInteractive(inst); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to connect to SSH: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}