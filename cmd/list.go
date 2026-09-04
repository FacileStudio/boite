package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all development sandboxes",
	Long:  `List all development sandboxes managed by boite.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkCommand("multipass"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n")
			fmt.Fprintln(os.Stderr, "Install: sudo snap install multipass")
			os.Exit(1)
		}

		output, err := RunCommand("multipass", "list")
		if err != nil {
			if strings.Contains(string(output), "permission denied") || strings.Contains(string(output), "Access denied") {
				fmt.Fprintf(os.Stderr, "Error: Permission denied accessing Multipass\n\n")
				fmt.Fprintln(os.Stderr, "Try with sudo:")
				fmt.Fprintln(os.Stderr, "  sudo boite list")
				fmt.Fprintln(os.Stderr)
				fmt.Fprintln(os.Stderr, "Or add your user to the multipass group:")
				fmt.Fprintln(os.Stderr, "  sudo usermod -aG multipass $USER")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error listing VMs: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(output))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
