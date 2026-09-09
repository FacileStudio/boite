package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start an existing sandbox",
	Long:  `Start a previously created sandbox VM.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		configFlag := cmd.Flag("config")
		configPath := ""
		if configFlag != nil {
			configPath = configFlag.Value.String()
		}
		qemu.ProgressStart()
		qemu.ProgressPhase("Starting VM")
		_, err := qemu.Start(name, configPath)
		qemu.ProgressStop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to start sandbox '%s': %v\n", name, err)
			os.Exit(1)
		}
		printSuccess(fmt.Sprintf("Sandbox '%s' started", name))
		printInfo(fmt.Sprintf("Connect: boite run %s", name))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
