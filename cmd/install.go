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
	Long: `Install boite to ~/.local/bin/ for global access.

Prerequisites:
  Multipass is required for VM-based sandboxes.
  Install via: sudo snap install multipass (Ubuntu) or brew install --cask multipass (macOS)`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
			os.Exit(1)
		}

		installDir := "/usr/local/bin"
		if !isRoot() {
			userLocalBin := filepath.Join(homeDir(), ".local", "bin")
			if err := os.MkdirAll(userLocalBin, 0o755); err == nil {
				installDir = userLocalBin
			}
		}

		destPath := filepath.Join(installDir, "boite")
		copyCmd := exec.Command("cp", exePath, destPath)
		output, err := copyCmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error installing boite: %v\n%s\n\n", err, string(output))
			fmt.Fprintln(os.Stderr, "Try with sudo:")
			fmt.Fprintf(os.Stderr, "  sudo %s install\n", exePath)
			os.Exit(1)
		}

		execCmd := exec.Command("chmod", "+x", destPath)
		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to chmod +x %s: %v\n", destPath, err)
		}

		fmt.Printf("boite installed to %s\n", destPath)
		fmt.Println("You can now run 'boite' from anywhere")

		pathEnv := os.Getenv("PATH")
		if !strings.Contains(pathEnv, installDir) {
			fmt.Fprintf(os.Stderr, "\nNote: Add to your PATH:\n")
			fmt.Fprintf(os.Stderr, "  export PATH=\"%s:$PATH\"\n", installDir)
			fmt.Fprintln(os.Stderr, "Add this to your ~/.zshrc or ~/.bashrc")
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
