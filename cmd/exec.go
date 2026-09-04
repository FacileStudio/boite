package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [name] [command...]",
	Short: "Execute a command in the sandbox VM",
	Long:  `Execute a command or open an interactive shell in the specified sandbox VM.`,
	Run: func(cmd *cobra.Command, args []string) {
		name := "dev-box"
		if len(args) > 0 {
			name = args[0]
		}

		if len(args) <= 1 {
			if err := runShell(name); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if err := ensureVMRunning(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		execArgs := append([]string{"multipass", "exec", name, "--"}, args[1:]...)
		if err := runInteractive(execArgs...); err != nil {
			if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "Access denied") {
				fmt.Fprintf(os.Stderr, "Error: Permission denied accessing VM '%s'\n", name)
			} else {
				fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
			}
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
