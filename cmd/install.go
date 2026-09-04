package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install boite to system PATH",
	Long:  `Install boite to ~/.local/bin/ for global access.

Prerequisites:
  Multipass is required for VM-based sandboxes.
  Install via: sudo snap install multipass (Ubuntu) or brew install --cask multipass (macOS)`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the current executable path
		exePath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
			os.Exit(1)
		}

		// Determine install directory
		installDir := "/usr/local/bin"
		if !isRoot() {
			// Try user local bin first
			userLocalBin := filepath.Join(homeDir(), ".local", "bin")
			if _, err := os.Stat(userLocalBin); err == nil {
				installDir = userLocalBin
			} else {
				// Create it if it doesn't exist
				if err := os.MkdirAll(userLocalBin, 0755); err == nil {
					installDir = userLocalBin
				}
			}
		}

		// Copy binary
		destPath := filepath.Join(installDir, "boite")
		copyCmd := exec.Command("cp", exePath, destPath)
		output, err := copyCmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error installing boite: %v\n%s\n", err, string(output))
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Try with sudo:")
			fmt.Fprintf(os.Stderr, "  sudo %s install\n", exePath)
			os.Exit(1)
		}

		// Make executable
		execCmd := exec.Command("chmod", "+x", destPath)
		execCmd.Run()

		fmt.Printf("boite installed to %s\n", destPath)
		fmt.Printf("You can now run 'boite' from anywhere\n")

		// Check if PATH is updated
		pathEnv := os.Getenv("PATH")
		if !strings.Contains(pathEnv, installDir) {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Note: Add to your PATH:")
			fmt.Fprintf(os.Stderr, "  export PATH=\"%s:$PATH\"\n", installDir)
			fmt.Fprintln(os.Stderr, "Add this to your ~/.zshrc or ~/.bashrc")
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}