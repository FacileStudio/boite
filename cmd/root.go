package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"
)

const asciiBanner = `▄▄▄▄   ▄▄▄  ▄▄ ▄▄▄▄▄▄ ▄▄▄▄▄
██▄██ ██▀██ ██   ██   ██▄▄
██▄█▀ ▀███▀ ██   ██   ██▄▄▄`

// version is written once by the linker (-ldflags -X ...cmd.version=...) at
// build time and only read afterwards.
var version = "0.7.2"

// Execute runs the root command, dispatching to the registered subcommands.
// fang renders the styled error and returns it, so a non-nil result is only
// forwarded as the exit code — the error is never printed a second time.
func Execute() error {
	if fang.Execute(context.Background(), newRootCmd(), fang.WithVersion(version)) != nil {
		os.Exit(1)
	}
	return nil
}

// newRootCmd assembles the boite command tree and its persistent flags.
func newRootCmd() *cobra.Command {
	var cfgFile string
	root := &cobra.Command{
		Use:     "boite",
		Short:   "Development sandbox manager",
		Long:    `A CLI tool to create, manage, and clean up virtual machine-based development sandboxes for general development.`,
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cfgFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		},
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.boite.yml)")
	root.AddCommand(rootSubcommands()...)
	return root
}

// rootSubcommands returns every boite subcommand in registration order.
func rootSubcommands() []*cobra.Command {
	return []*cobra.Command{
		newCreateCmd(),
		newRunCmd(),
		newListCmd(),
		newStartCmd(),
		newStopCmd(),
		newRmCmd(),
		newExecCmd(),
		newSyncCmd(),
		newEnvCmd(),
		newSnapshotCmd(),
		newRollbackCmd(),
		newSnapshotsCmd(),
	}
}

// versionString returns the semantic version with any leading "v" stripped.
func versionString() string {
	v := strings.TrimPrefix(version, "v")
	if v == "" || v == "dev" {
		return version
	}
	return v
}
