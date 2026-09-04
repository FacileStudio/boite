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

		output, err := RunCommand("multipass", "list")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking VMs: %v\n", err)
			os.Exit(1)
		}

		if strings.Contains(string(output), name) {
			startExistingVM(name)
		} else {
			createNewVM(name, noMount)
		}

		fmt.Printf("VM: %s\n", name)
		fmt.Println("User: boite (/home/boite)")
		if !noMount {
			fmt.Printf("Workspace: /workspace\n")
		}
		fmt.Println("To access the VM, run: boite shell", name)
	},
}

func startExistingVM(name string) {
	err := Spinner(fmt.Sprintf("Starting VM '%s'", name), func() error {
		_, err := RunCommand("multipass", "start", name)
		return err
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting VM: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("VM '%s' started\n", name)
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
	cloudInitPath := viper.GetString("cloud_init_path")
	if cloudInitPath == "" {
		fmt.Fprintln(os.Stderr, "Error: cloud_init_path not set in config")
		os.Exit(1)
	}

	fmt.Printf("Creating VM '%s'...\n", name)
	fmt.Println("This may take a few minutes to provision...")
	fmt.Println("=== Cloud-init output ===")

	launchArgs := vmLaunchArgs(name, cloudInitPath, workspacePath, noMount)
	err := runInteractive(append([]string{"multipass", "launch"}, launchArgs...)...)
	fmt.Println("=== End cloud-init output ===")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating VM: %v\n\n", err)
		fmt.Fprintln(os.Stderr, "Try with sudo:")
		fmt.Fprintf(os.Stderr, "  sudo multipass launch 24.04 --name %s --cloud-init %s", name, cloudInitPath)
		if !noMount {
			fmt.Fprintf(os.Stderr, " --mount %s:/workspace", workspacePath)
		}
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	if !noMount {
		syncWorkspaceZshrc(name)
	}

	fmt.Printf("VM '%s' created and started\n", name)
}

func syncWorkspaceZshrc(name string) {
	cmdStr := "if [ -f /workspace/.zshrc_local ]; then " +
		"if id boite >/dev/null 2>&1; then " +
		"cp /workspace/.zshrc_local /home/boite/.zshrc && chown boite:boite /home/boite/.zshrc; " +
		"fi; fi"
	RunCommand("multipass", "exec", name, "--", "sudo", "bash", "-c", cmdStr)
}

func init() {
	createCmd.Flags().Bool("no-mount", false, "Create VM without mounting workspace")
	rootCmd.AddCommand(createCmd)
}
