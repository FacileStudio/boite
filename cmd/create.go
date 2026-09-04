package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new development sandbox (VM)",
	Long:  `Create a new development sandbox VM with development tools pre-installed.
Use --no-mount to create a VM without mounting the current workspace (like the old 'build' command).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noMount, _ := cmd.Flags().GetBool("no-mount")
		name := args[0]
		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: sandbox name is required")
			os.Exit(1)
		}

		// Check if multipass is available
		if err := checkCommand("multipass"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: multipass not found in PATH\n")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Install Multipass:")
			fmt.Fprintln(os.Stderr, "  Ubuntu: sudo snap install multipass")
			fmt.Fprintln(os.Stderr, "  macOS: brew install --cask multipass")
			fmt.Fprintln(os.Stderr, "  Windows: winget install Canonical.Multipass")
			os.Exit(1)
		}

		// Get current user for username mapping
		currentUser, err := user.Current()
		if err != nil {
			currentUser = &user.User{Uid: "1000", Gid: "1000", Username: "devuser", HomeDir: "/home/devuser"}
		}

		// Check if VM exists
		output, err := RunCommand("multipass", "list")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking VMs: %v\n", err)
			os.Exit(1)
		}

		vmExists := false
		if strings.Contains(string(output), name) {
			vmExists = true
		}

		if vmExists {
			// VM exists, start it
			err := Spinner(fmt.Sprintf("Starting VM '%s'", name), func() error {
				_, err := RunCommand("multipass", "start", name)
				return err
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error starting VM: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("VM '%s' started\n", name)
		} else {
			// Create new VM
			workspacePath, _ := os.Getwd()

			// Get cloud-init path from config (set by root.go)
			cloudInitPath := viper.GetString("cloud_init_path")
			if cloudInitPath == "" {
				fmt.Fprintln(os.Stderr, "Error: cloud_init_path not set in config")
				os.Exit(1)
			}
			// Debug: write a copy to home for inspection
			debugPath := filepath.Join(homeDir(), "cloud-init-debug.yml")
			input, err := os.ReadFile(cloudInitPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading cloud-init for debug: %v\n", err)
			} else {
				if err := os.WriteFile(debugPath, input, 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing debug cloud-init: %v\n", err)
				}
			}
			// Log the temp file path
			fmt.Fprintf(os.Stderr, "Using cloud-init temp file: %s\n", cloudInitPath)

			// Get VM config
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

			fmt.Printf("Creating VM '%s'...\n", name)
			fmt.Println("This may take a few minutes to provision...")

			launchArgs := []string{
				"24.04",
				"--name", name,
				"--memory", memory,
				"--disk", disk,
				"--cpus", fmt.Sprintf("%d", cpus),
				"--cloud-init", cloudInitPath,
			}
			if !noMount {
				launchArgs = append(launchArgs, "--mount", workspacePath+":/workspace")
			}

			spinnerErr := Spinner(fmt.Sprintf("Creating VM '%s'", name), func() error {
				_, errInner := RunCommand(append([]string{"multipass", "launch"}, launchArgs...)...)
				if errInner != nil {
					return fmt.Errorf("%v", errInner)
				}
				return nil
			})
			if spinnerErr != nil {
				fmt.Fprintf(os.Stderr, "Error creating VM: %v\n", spinnerErr)
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, "Try with sudo:")
				fmt.Fprintf(os.Stderr, "  sudo multipass launch 24.04 --name %s --cloud-init %s", name, cloudInitPath)
				if !noMount {
					fmt.Fprintf(os.Stderr, " --mount %s:/workspace", workspacePath)
				}
				fmt.Fprintln(os.Stderr)
				os.Exit(1)
			}
			fmt.Printf("VM '%s' created and started\n", name)
		}

		fmt.Printf("VM: %s\n", name)
		fmt.Printf("User: %s\n", currentUser.Username)
		if !noMount {
			fmt.Printf("Workspace: /workspace\n")
		}
		fmt.Println("To access the VM, run: boite shell", name)
	},
}

func init() {
	createCmd.Flags().Bool("no-mount", false, "Create VM without mounting the current workspace (like the old 'build' command)")
	rootCmd.AddCommand(createCmd)
}