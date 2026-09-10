package qemu

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// VMConfig represents the vm section of ~/.boite.yml
type VMConfig struct {
	CPUs   int    `yaml:"cpus"`
	Disk   string `yaml:"disk"`
	Memory string `yaml:"memory"`
}

// WorkspaceConfig represents the workspace section of ~/.boite.yml. SyncAtRun
// is a pointer so an absent setting can be told apart from an explicit false;
// sync is an opt-in, so both mean "disabled".
type WorkspaceConfig struct {
	SyncAtRun *bool `yaml:"sync_at_run"`
}

// WorkspaceSyncEnabled reports whether the current directory should be synced
// into the sandbox's /workspace at run time. Sync is off by default and must
// be opted into with workspace.sync_at_run: true. The instance's own
// --no-mount setting and an explicit run override both disable it; any single
// "no" wins, and the run override still applies to an opted-in instance.
func WorkspaceSyncEnabled(inst *Instance, configPath string, forceSkip bool) bool {
	if inst.NoMount || forceSkip {
		return false
	}
	cfg, err := LoadBoiteConfig(configPath)
	if err == nil && cfg != nil && cfg.Workspace != nil && cfg.Workspace.SyncAtRun != nil {
		return *cfg.Workspace.SyncAtRun
	}
	return false
}

// BoiteConfig represents the top-level ~/.boite.yml structure. Toolchain,
// packages, users and dotfiles live in the baked base image, so the config
// only tunes VM and workspace behaviour.
type BoiteConfig struct {
	VM        *VMConfig        `yaml:"vm"`
	Workspace *WorkspaceConfig `yaml:"workspace"`
}

// LoadBoiteConfig reads and parses a boite YAML config file. If path is empty,
// it falls back to ~/.boite.yml. A missing file returns nil, nil so callers can
// fall back to defaults.
func LoadBoiteConfig(path string) (*BoiteConfig, error) {
	configPath := path
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		configPath = filepath.Join(home, ".boite.yml")
	}

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat config: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg BoiteConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}
