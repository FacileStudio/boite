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

// Sync of the current directory into the sandbox at run time is controlled by
// the top-level sync key of ~/.boite.yml. It is a pointer so an absent setting
// can be told apart from an explicit false; sync is an opt-in, so both mean
// "disabled".
//
// WorkspaceSyncEnabled reports whether the current directory should be synced
// into the sandbox's /workspace at run time. Sync is off by default and must
// be opted into with sync: true. The instance's own --no-mount setting and an
// explicit run override both disable it; any single "no" wins, and the run
// override still applies to an opted-in instance.
func WorkspaceSyncEnabled(inst *Instance, configPath string, forceSkip bool) bool {
	if inst.NoMount || forceSkip {
		return false
	}
	cfg, err := LoadBoiteConfig(configPath)
	if err == nil && cfg != nil && cfg.Sync != nil {
		return *cfg.Sync
	}
	return false
}

// ProvisionConfig is the per-instance provisioning block of ~/.boite.yml.
// Packages are apt packages installed into the guest, and commands are shell
// lines run as the boite user. The base toolchain stays baked in the image;
// this tunes what each new instance adds on top of it.
type ProvisionConfig struct {
	Packages []string `yaml:"packages"`
	Commands []string `yaml:"commands"`
}

// Empty reports whether the provision block declares nothing to do.
func (p *ProvisionConfig) Empty() bool {
	return len(p.Packages) == 0 && len(p.Commands) == 0
}

// BoiteConfig represents the top-level ~/.boite.yml structure. Toolchain,
// packages, users and dotfiles live in the baked base image, so the config
// only tunes VM, sync and per-instance provision behaviour.
type BoiteConfig struct {
	VM        *VMConfig        `yaml:"vm"`
	Sync      *bool            `yaml:"sync"`
	Provision *ProvisionConfig `yaml:"provision"`
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
