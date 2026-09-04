package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
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
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n")
			os.Exit(1)
		}

		_ = Spinner(fmt.Sprintf("Stopping VM '%s'", name), func() error {
			output, err := RunCommand("multipass", "stop", name)
			if err != nil {
				return fmt.Errorf("%v\n%s", err, string(output))
			}
			return nil
		})

		fmt.Printf("VM '%s' stopped\n", name)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}