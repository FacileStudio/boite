package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/lipgloss/v2"
	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newCreateCmd builds the "create" subcommand.
func newCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new development sandbox",
		Long: `Create a new development sandbox VM with development tools pre-installed.
Uses cached Debian 13 image with Copy-on-Write overlay.
Use --no-mount so 'boite run' never syncs a workspace into /workspace for this instance.`,
		Args: cobra.ExactArgs(1),
		Run:  runCreate,
	}
	createCmd.Flags().Bool("no-mount", false, "Create VM without mounting workspace")
	createCmd.Flags().Bool("generate-key", false, "Generate a unique SSH key pair for this VM instead of using your existing ~/.ssh key")
	return createCmd
}

func runCreate(cmd *cobra.Command, args []string) {
	noMount, _ := cmd.Flags().GetBool("no-mount")
	generateKey, _ := cmd.Flags().GetBool("generate-key")
	cfgPath := configPath(cmd)
	name := args[0]

	if name == "" {
		fmt.Fprintln(os.Stderr, "Error: sandbox name is required")
		os.Exit(1)
	}

	workspacePath, _ := os.Getwd()

	if !isConfigPresent(cfgPath) {
		printInfo("No ~/.boite.yml file found nor config file passed, falling back to default config")
	}

	fmt.Fprintf(os.Stderr, "Creating sandbox '%s'...\n", name)

	qemu.ProgressStart()
	inst, err := qemu.Create(name, workspacePath, noMount, cfgPath, generateKey)
	qemu.ProgressStop()

	if err != nil {
		printError(fmt.Sprintf("Failed to create sandbox '%s': %v", name, err))
		os.Exit(1)
	}

	fmt.Println()
	lipgloss.Println(renderSandboxCard(name, workspacePath, noMount, inst.SSHPort))
}

// isConfigPresent reports whether a config file is available for this run.
func isConfigPresent(cfgPath string) bool {
	if cfgPath != "" {
		return true
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(home, ".boite.yml"))
	return err == nil
}
