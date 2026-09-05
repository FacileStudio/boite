package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new development sandbox",
	Long: `Create a new development sandbox VM with development tools pre-installed.
Use --no-mount to create a VM without mounting the current workspace.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noMount, _ := cmd.Flags().GetBool("no-mount")
		name := args[0]
		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: sandbox name is required")
			os.Exit(1)
		}

		if err := checkCommand("multipass"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n\n")
			fmt.Fprintln(os.Stderr, "Install Multipass:")
			fmt.Fprintln(os.Stderr, "  Ubuntu: sudo snap install multipass")
			fmt.Fprintln(os.Stderr, "  macOS: brew install --cask multipass")
			fmt.Fprintln(os.Stderr, "  Windows: winget install Canonical.Multipass")
			os.Exit(1)
		}

		output, err := runCommand("multipass", "list")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking VMs: %v\n", err)
			os.Exit(1)
		}

		if strings.Contains(string(output), name) {
			startExistingVM(name)
		} else {
			createNewVM(name, noMount)
		}
	},
}

func startExistingVM(name string) {
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
}

func vmLaunchArgs(name, cloudInitPath, workspacePath string, noMount bool) []string {
	memory := viper.GetString("vm.memory")
	if memory == "" {
		memory = "4G"
	}
	disk := viper.GetString("vm.disk")
	if disk == "" {
		disk = "40G"
	}
	cpus := viper.GetInt("vm.cpus")
	if cpus == 0 {
		cpus = 2
	}

	args := []string{
		"24.04",
		"--name", name,
		"--memory", memory,
		"--disk", disk,
		"--cpus", fmt.Sprintf("%d", cpus),
		"--cloud-init", cloudInitPath,
	}
	if !noMount {
		args = append(args, "--mount", workspacePath+":/workspace")
	}
	return args
}

func createNewVM(name string, noMount bool) {
	workspacePath, _ := os.Getwd()
	cloudInitPath, err := ensureCloudInit(homeDir())
	if err != nil {
		printError(fmt.Sprintf("Error preparing cloud-init: %v", err))
		os.Exit(1)
	}

	printInfo(fmt.Sprintf("Creating sandbox '%s'...", name))
	printInfo("Provisioning environment via cloud-init...")

	launchArgs := vmLaunchArgs(name, cloudInitPath, workspacePath, noMount)
	err = runInteractive(append([]string{"multipass", "launch"}, launchArgs...)...)
	if err != nil {
		printError(fmt.Sprintf("Failed to create sandbox '%s': %v", name, err))
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access denied") {
			printInfo("Permission denied accessing Multipass. Try adding your user to the multipass group:")
			fmt.Fprintf(os.Stderr, "  sudo usermod -aG multipass $USER\n")
		}
		os.Exit(1)
	}

	if !noMount {
		syncWorkspaceZshrc(name)
	}

	fmt.Println()
	fmt.Println(renderSandboxCard(name, workspacePath, noMount))
}

func init() {
	createCmd.Flags().Bool("no-mount", false, "Create VM without mounting workspace")
	rootCmd.AddCommand(createCmd)
}
