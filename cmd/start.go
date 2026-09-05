package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a development sandbox VM",
	Long:  `Start a previously created or stopped sandbox VM.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := checkCommand("multipass"); err != nil {
			printError("multipass not found in PATH")
			os.Exit(1)
		}

		err := spinner(fmt.Sprintf("Starting sandbox '%s'", name), func() error {
			_, err := runCommand("multipass", "start", name)
			return err
		})
		if err != nil {
			printError(fmt.Sprintf("Failed to start sandbox '%s': %v", name, err))
			os.Exit(1)
		}
		printSuccess(fmt.Sprintf("Sandbox '%s' started", name))
		printInfo(fmt.Sprintf("Connect: boite shell %s", name))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
