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
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n")
			os.Exit(1)
		}

		RunCommand("multipass", "stop", name)

		err := Spinner(fmt.Sprintf("Removing VM '%s'", name), func() error {
			output, err := RunCommand("multipass", "delete", "--purge", name)
			if err != nil {
				return fmt.Errorf("%v\n%s", err, string(output))
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("VM '%s' removed\n", name)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
