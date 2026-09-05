package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a development sandbox VM",
	Long:  `Stop the specified sandbox VM. The data is preserved.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "dev-box"
		if len(args) > 0 {
			name = args[0]
		}

		if err := checkCommand("multipass"); err != nil {
			printError("multipass not found in PATH")
			os.Exit(1)
		}

		err := spinner(fmt.Sprintf("Stopping sandbox '%s'", name), func() error {
			output, err := runCommand("multipass", "stop", name)
			if err != nil {
				return fmt.Errorf("%v\n%s", err, string(output))
			}
			return nil
		})
		if err != nil {
			printError(fmt.Sprintf("Failed to stop sandbox '%s': %v", name, err))
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Sandbox '%s' stopped", name))
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
