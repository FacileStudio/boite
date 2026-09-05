package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove a development sandbox VM",
	Long:    `Remove the specified sandbox VM. This will delete the VM permanently.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "dev-box"
		if len(args) > 0 {
			name = args[0]
		}

		if err := checkCommand("multipass"); err != nil {
			printError("multipass not found in PATH")
			os.Exit(1)
		}

		runCommand("multipass", "stop", name)

		err := spinner(fmt.Sprintf("Removing sandbox '%s'", name), func() error {
			output, err := runCommand("multipass", "delete", "--purge", name)
			if err != nil {
				return fmt.Errorf("%v\n%s", err, string(output))
			}
			return nil
		})
		if err != nil {
			printError(fmt.Sprintf("Failed to remove sandbox '%s': %v", name, err))
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Sandbox '%s' removed", name))
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
