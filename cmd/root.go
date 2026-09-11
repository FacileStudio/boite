package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const asciiBanner = `▄▄▄▄   ▄▄▄  ▄▄ ▄▄▄▄▄▄ ▄▄▄▄▄
██▄██ ██▀██ ██   ██   ██▄▄
██▄█▀ ▀███▀ ██   ██   ██▄▄▄`

// version is written once by the linker (-ldflags -X ...cmd.version=...) at
// build time and only read afterwards.
var version = "0.7.0"

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
	root.AddCommand(
		newCreateCmd(),
		newRunCmd(),
		newListCmd(),
		newStartCmd(),
		newStopCmd(),
		newRmCmd(),
		newExecCmd(),
		newSyncCmd(),
		newEnvCmd(),
	)
	return root
}

// versionString returns the semantic version with any leading "v" stripped.
func versionString() string {
	v := strings.TrimPrefix(version, "v")
	if v == "" || v == "dev" {
		return version
	}
	return v
}

// configPath resolves the value of the persistent --config flag.
func configPath(cmd *cobra.Command) string {
	flag := cmd.Flag("config")
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

func initConfig(cfgFile string) error {
	home := homeDir()
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(home)
		viper.SetConfigType("yml")
		viper.SetConfigName(".boite")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, configNotFound := err.(viper.ConfigFileNotFoundError); !configNotFound {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if cfgFile == "" {
			return createDefaultConfig(home)
		}
	}

	return nil
}

func createDefaultConfig(home string) error {
	configPath := filepath.Join(home, ".boite.yml")
	defaultConfig := map[string]any{
		"vm": map[string]any{
			"cpus":   2,
			"memory": "2G",
			"disk":   "20G",
		},
	}
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}
	printSuccess(fmt.Sprintf("Created default config at %s", configPath))
	return nil
}
