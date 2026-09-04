package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:     "shell [name]",
	Aliases: []string{"ssh", "enter"},
	Short:   "Open an interactive shell in the sandbox VM",
	Long:    `Open an interactive shell session in the specified sandbox VM.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "dev-box"
		if len(args) > 0 {
			name = args[0]
		}

		if err := runShell(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runShell(name string) error {
	if err := ensureVMRunning(name); err != nil {
		return err
	}
	fmt.Println("Connecting to VM... Exit with 'exit' or Ctrl+D.")
	if err := runInteractive("multipass", "shell", name); err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "Access denied") {
			return fmt.Errorf("permission denied accessing VM '%s'. Try with sudo: sudo boite shell %s", name, name)
		}
		return fmt.Errorf("connecting to VM: %w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
