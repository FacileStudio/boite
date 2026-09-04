package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a development sandbox VM",
	Long:  `Start a previously created or stopped sandbox VM.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := checkCommand("multipass"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n")
			os.Exit(1)
		}

		err := Spinner(fmt.Sprintf("Starting VM '%s'", name), func() error {
			_, err := RunCommand("multipass", "start", name)
			return err
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting VM: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("VM '%s' started\n", name)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}