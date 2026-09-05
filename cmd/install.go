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
			printError(fmt.Sprintf("Failed to install boite: %v\n%s", err, string(output)))
			printInfo(fmt.Sprintf("If installing to /usr/local/bin, try: sudo %s install", exePath))
			os.Exit(1)
		}

		execCmd := exec.Command("chmod", "+x", destPath)
		if err := execCmd.Run(); err != nil {
			printError(fmt.Sprintf("Failed to chmod +x %s: %v", destPath, err))
		}

		printSuccess(fmt.Sprintf("boite installed to %s", destPath))
		printInfo("You can now run 'boite' from anywhere")

		pathEnv := os.Getenv("PATH")
		if !strings.Contains(pathEnv, installDir) {
			printInfo(fmt.Sprintf("Add to your PATH: export PATH=\"%s:$PATH\"", installDir))
		}
	},
}

func isRoot() bool {
	return os.Geteuid() == 0
}

func init() {
	rootCmd.AddCommand(installCmd)
}
