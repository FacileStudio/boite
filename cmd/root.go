package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const asciiBanner = `▄▄
██          ▀▀  ██
████▄ ▄███▄ ██ ▀██▀▀ ▄█▀█▄
██ ██ ██ ██ ██  ██   ██▄█▀
████▀ ▀███▀ ██▄ ██   ▀█▄▄▄ `

var (
	Version = "0.1.7"
	version = ""
)
var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "boite",
	Short: "Development sandbox manager",
	Long: `A CLI tool to create, manage, and clean up virtual machine-based development sandboxes for general development.

Requires: Multipass (https://multipass.run/) — install via:
  Ubuntu: sudo snap install multipass
  macOS:  brew install --cask multipass
  Windows: winget install Canonical.Multipass`,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func versionString() string {
	v := strings.TrimPrefix(Version, "v")
	if v == "" || v == "dev" {
		return Version
	}
	return "v" + v
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
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if err := createDefaultConfig(home); err != nil {
			return err
		}
	}

	return nil
}

func createDefaultConfig(home string) error {
	configPath := filepath.Join(home, ".boite.yml")
	var cloudInitObj map[string]interface{}
	if err := yaml.Unmarshal([]byte(defaultCloudInit()), &cloudInitObj); err != nil {
		return fmt.Errorf("failed to parse default cloud-init: %w", err)
	}
	defaultConfig := map[string]interface{}{
		"vm": map[string]interface{}{
			"cpus":   2,
			"memory": "4G",
			"disk":   "40G",
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

func ensureCloudInit(home string) (string, error) {
	var cloudInitStr string
	val := viper.Get("cloud_init")
	switch v := val.(type) {
	case string:
		cloudInitStr = v
	case map[string]interface{}:
		b, err := yaml.Marshal(v)
		if err != nil {
			cloudInitStr = defaultCloudInit()
		} else {
			cloudInitStr = "#cloud-config\n" + string(b)
		}
	case map[interface{}]interface{}:
		b, err := yaml.Marshal(v)
		if err != nil {
			cloudInitStr = defaultCloudInit()
		} else {
			cloudInitStr = "#cloud-config\n" + string(b)
		}
	default:
		legacyPath := filepath.Join(home, ".cloud-init.yml")
		b, err := os.ReadFile(legacyPath)
		if err != nil {
			cloudInitStr = defaultCloudInit()
		} else {
			cloudInitStr = string(b)
		}
	}

	cloudInitStr = strings.TrimSpace(cloudInitStr) + "\n"
	if !strings.HasPrefix(cloudInitStr, "#cloud-config") {
		cloudInitStr = "#cloud-config\n" + cloudInitStr
	}

	var testNode yaml.Node
	if err := yaml.Unmarshal([]byte(cloudInitStr), &testNode); err != nil {
		cloudInitStr = defaultCloudInit()
	}

	dataDir := filepath.Join(home, ".boite")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create data dir: %w", err)
	}
	cloudInitPath := filepath.Join(dataDir, "cloud-init.yml")
	if err := os.WriteFile(cloudInitPath, []byte(cloudInitStr), 0o644); err != nil {
		return "", fmt.Errorf("failed to write cloud-init to data file: %w", err)
	}
	return cloudInitPath, nil
}
