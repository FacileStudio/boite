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
		jsonOutput, err := runCommand("multipass", "list", "--format", "json")
		if err != nil {
			errText := strings.ToLower(string(jsonOutput))
			if strings.Contains(errText, "permission denied") || strings.Contains(errText, "access denied") {
				printError("Permission denied accessing Multipass")
				printInfo("Add your user to the multipass group:")
				fmt.Fprintln(os.Stderr, "  sudo usermod -aG multipass $USER")
				os.Exit(1)
			}
			printError(fmt.Sprintf("Failed to list sandboxes: %v", err))
			os.Exit(1)
		}

		if rendered, ok := renderInstanceTable(jsonOutput); ok {
			fmt.Println(rendered)
			return
		}

		fmt.Print(string(jsonOutput))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
