package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new development sandbox",
	Long: `Create a new development sandbox VM with development tools pre-installed.
Uses cached Debian 13 image with Copy-on-Write overlay.
Use --no-mount to create a VM without mounting the current workspace.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noMount, _ := cmd.Flags().GetBool("no-mount")
		generateKey, _ := cmd.Flags().GetBool("generate-key")
		name := args[0]

		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: sandbox name is required")
			os.Exit(1)
		}

		workspacePath, _ := os.Getwd()

		if !isConfigPresent() {
			printInfo("No ~/.boite.yml file found nor config file passed, falling back to default config")
		}

		fmt.Fprintln(os.Stderr, fmt.Sprintf("Creating sandbox '%s'...", name))

		qemu.ProgressStart()
		inst, err := qemu.Create(name, workspacePath, noMount, cfgFile, generateKey)
		qemu.ProgressStop()

		if err != nil {
			printError(fmt.Sprintf("Failed to create sandbox '%s': %v", name, err))
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println(renderSandboxCard(name, workspacePath, noMount, inst.SSHPort))
	},
}

func init() {
	createCmd.Flags().Bool("no-mount", false, "Create VM without mounting workspace")
	createCmd.Flags().Bool("generate-key", false, "Generate a unique SSH key pair for this VM instead of using your existing ~/.ssh key")
	rootCmd.AddCommand(createCmd)
}

func isConfigPresent() bool {
	if cfgFile != "" {
		return true
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(home, ".boite.yml"))
	return err == nil
}
