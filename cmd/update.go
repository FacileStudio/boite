package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update Boite VM to the latest version",
	Long: `Update a running Boite VM to the latest available version.
If no VM name is provided, updates the currently active shell VM (if any).
The update process checks GitHub for the latest release and replaces the boite binary.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var vmName string
		if len(args) == 1 {
			vmName = args[0]
		} else {
			// Try to get current shell VM
			vmName = getCurrentShellVM()
			if vmName == "" {
				fmt.Fprintln(os.Stderr, "Error: no VM specified and no active shell VM found")
				fmt.Fprintln(os.Stderr, "Usage: boite update <vm-name>")
				os.Exit(1)
			}
		}

		if err := updateVM(vmName); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating VM: %v\n", err)
			os.Exit(1)
		}
	},
}

// getCurrentShellVM tries to detect which VM we're currently shell'd into
// by checking the MULTIPASS_INSTANCE_NAME environment variable (set by boite shell)
func getCurrentShellVM() string {
	return os.Getenv("MULTIPASS_INSTANCE_NAME")
}

// updateVM performs the update process on the specified VM
func updateVM(name string) error {
	fmt.Printf("Checking for updates for VM '%s'...\n", name)

	// First, check if VM exists and is running
	if err := ensureVMRunning(name); err != nil {
		return err
	}

	// Get latest version from GitHub
	latestVersion, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Get current version inside VM
	currentVersion, err := getVMVersion(name)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == latestVersion {
		fmt.Printf("VM '%s' is already at the latest version (%s)\n", name, currentVersion)
		return nil
	}

	fmt.Printf("Updating VM '%s' from %s to %s...\n", name, currentVersion, latestVersion)

	// Download and install the update
	if err := performUpdate(name, latestVersion); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// Verify the update
	newVersion, err := getVMVersion(name)
	if err != nil {
		return fmt.Errorf("failed to verify update: %w", err)
	}

	if newVersion != latestVersion {
		return fmt.Errorf("update verification failed: expected %s, got %s", latestVersion, newVersion)
	}

	fmt.Printf("VM '%s' successfully updated to %s\n", name, newVersion)
	return nil
}

// getLatestRelease fetches the latest release tag from GitHub
func getLatestRelease() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/FacileStudio/boite/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	// Remove 'v' prefix if present for comparison
	version := strings.TrimPrefix(release.TagName, "v")
	if version == "" {
		version = release.TagName
	}
	return version, nil
}

// getVMVersion runs 'boite --version' inside the VM and returns the version
func getVMVersion(name string) (string, error) {
	output, err := RunCommand("multipass", "exec", name, "--", "boite", "--version")
	if err != nil {
		return "", err
	}
	// Output is like "boite v0.1.5"
	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected version output: %s", string(output))
	}
	version := strings.TrimPrefix(parts[1], "v")
	return version, nil
}

// performUpdate downloads and installs the update script in the VM
func performUpdate(name string, version string) error {
	// Stream the install script directly into the VM
	// We'll use a heredoc to avoid needing to quote the script
	updateScript := fmt.Sprintf(`
set -e
cd /tmp
curl -fsSL https://raw.githubusercontent.com/FacileStudio/boite/main/install.sh | BOITE_VERSION=%s bash
`, version)

	// Run the update script as boite user
	_, err := RunCommand("multipass", "exec", name, "--", "sudo", "-u", "boite", "bash", "-c", updateScript)
	return err
}

func init() {
	updateCmd.Flags().BoolP("help", "h", false, "Help for update")
	rootCmd.AddCommand(updateCmd)
}