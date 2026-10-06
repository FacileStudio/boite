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
	createCmd.Flags().StringP("distro", "d", "", "Guest distribution (debian, alpine, nixos)")
	createCmd.Flags().Bool("nixos", false, "Use NixOS guest distribution (shorthand for --distro nixos)")
	createCmd.Flags().Bool("alpine", false, "Use Alpine guest distribution (shorthand for --distro alpine)")
	createCmd.Flags().Bool("debian", false, "Use Debian guest distribution (shorthand for --distro debian)")
	return createCmd
}

func extractDistroFlag(cmd *cobra.Command) string {
	if n, _ := cmd.Flags().GetBool("nixos"); n {
		return "nixos"
	}
	if a, _ := cmd.Flags().GetBool("alpine"); a {
		return "alpine"
	}
	if d, _ := cmd.Flags().GetBool("debian"); d {
		return "debian"
	}
	val, _ := cmd.Flags().GetString("distro")
	return val
}

func runCreate(cmd *cobra.Command, args []string) {
	cfgPath := configPath(cmd)
	name := args[0]
	if name == "" {
		fmt.Fprintln(os.Stderr, "Error: sandbox name is required")
		os.Exit(1)
	}

	distro, err := resolveCreateDistro(extractDistroFlag(cmd), cfgPath)
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	if !isConfigPresent(cfgPath) {
		printInfo("No ~/.boite.yml file found nor config file passed, falling back to default config")
	}

	fmt.Fprintf(os.Stderr, "Creating sandbox '%s'...\n", name)
	opts := buildCreateOptions(cmd, name, distro, cfgPath)

	qemu.ProgressStart()
	inst, err := qemu.Create(opts)
	qemu.ProgressStop()

	if err != nil {
		printError(fmt.Sprintf("Failed to create sandbox '%s': %v", name, err))
		os.Exit(1)
	}

	fmt.Println()
	lipgloss.Println(renderSandboxCard(name, inst.Distro, opts.WorkspacePath, opts.NoMount, inst.SSHPort))
}

func buildCreateOptions(cmd *cobra.Command, name, distro, cfgPath string) qemu.CreateOptions {
	noMount, _ := cmd.Flags().GetBool("no-mount")
	generateKey, _ := cmd.Flags().GetBool("generate-key")
	workspacePath, _ := os.Getwd()
	return qemu.CreateOptions{
		Name:          name,
		WorkspacePath: workspacePath,
		NoMount:       noMount,
		ConfigPath:    cfgPath,
		GenerateKey:   generateKey,
		Distro:        distro,
	}
}

func resolveCreateDistro(flagVal, cfgPath string) (string, error) {
	if flagVal != "" {
		spec, err := qemu.GetDistro(flagVal)
		if err != nil {
			return "", err
		}
		return spec.Name, nil
	}
	cfg, err := qemu.LoadBoiteConfig(cfgPath)
	if err != nil {
		return "", err
	}
	if cfg != nil && cfg.VM != nil {
		return cfg.VM.EffectiveDistro(), nil
	}
	return qemu.DefaultDistro, nil
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
