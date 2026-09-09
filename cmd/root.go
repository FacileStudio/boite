package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const asciiBanner = `▄▄▄▄   ▄▄▄  ▄▄ ▄▄▄▄▄▄ ▄▄▄▄▄
██▄██ ██▀██ ██   ██   ██▄▄
██▄█▀ ▀███▀ ██   ██   ██▄▄▄`

var (
	Version = "0.1.13"
	version = ""
)
var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "boite",
	Short:   "Development sandbox manager",
	Long:    `A CLI tool to create, manage, and clean up virtual machine-based development sandboxes for general development.`,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	},
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
}

func Execute() error {
	return rootCmd.Execute()
}

func versionString() string {
	v := strings.TrimPrefix(Version, "v")
	if v == "" || v == "dev" {
		return Version
	}
	return v
}

func init() {
	if version != "" {
		Version = version
	}
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.boite.yml)")
}

func initConfig() error {
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
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if cfgFile != "" {

			} else {
				if err := createDefaultConfig(home); err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

func createDefaultConfig(home string) error {
	configPath := filepath.Join(home, ".boite.yml")
	var cloudInitObj map[string]any
	if err := yaml.Unmarshal([]byte(qemu.DefaultCloudInitYAML()), &cloudInitObj); err != nil {
		return fmt.Errorf("failed to parse default cloud-init: %w", err)
	}
	defaultConfig := map[string]any{
		"vm": map[string]any{
			"cpus":   2,
			"memory": "2G",
			"disk":   "20G",
		},
		"cloud_init": cloudInitObj,
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
