package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [name]",
	Aliases: []string{"ssh", "enter"},
	Short: "Open an interactive shell in the sandbox VM",
	Long:  `Open an interactive shell session in the specified sandbox VM.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "dev-box"
		if len(args) > 0 {
			name = args[0]
		}

		// Check if multipass is available
		if err := checkCommand("multipass"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n")
			os.Exit(1)
		}

		// Check if VM exists and is running
		output, err := RunCommand("multipass", "info", name)
		if err != nil {
			if strings.Contains(string(output), "not found") || strings.Contains(string(output), "no instance named") {
				fmt.Fprintf(os.Stderr, "Error: VM '%s' not found\n", name)
				fmt.Fprintln(os.Stderr, "Create it first: boite create", name)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error checking VM: %v\n", err)
			os.Exit(1)
		}

		// Check if VM is running
		if strings.Contains(string(output), "Status: Running") == false {
			fmt.Printf("VM '%s' is not running, starting...\n", name)
			_, err := RunCommand("multipass", "start", name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error starting VM: %v\n", err)
				os.Exit(1)
			}
		}

		// Open shell
		fmt.Println("Connecting to VM... Exit with 'exit' or Ctrl+D.")
		shellCmd := exec.Command("multipass", "shell", name)
		shellCmd.Stdin = os.Stdin
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr
		if err := shellCmd.Run(); err != nil {
			// Check for permission issues
			if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "Access denied") {
				fmt.Fprintf(os.Stderr, "\nError: Permission denied accessing VM '%s'\n", name)
				fmt.Fprintln(os.Stderr, "Try with sudo: sudo boite shell", name)
			} else {
				fmt.Fprintf(os.Stderr, "Error connecting to VM: %v\n", err)
			}
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}