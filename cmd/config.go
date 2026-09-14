package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

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
