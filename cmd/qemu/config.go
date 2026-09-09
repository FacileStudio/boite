package qemu

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed cloudinit.yml
var defaultCloudInitYAML string

// DefaultCloudInitYAML returns the embedded default cloud-init YAML content
func DefaultCloudInitYAML() string {
	return defaultCloudInitYAML
}

// WriteFileConfig represents a write_files entry in cloud-init
type WriteFileConfig struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Owner       string `yaml:"owner,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Encoding    string `yaml:"encoding,omitempty"`
	Defer       bool   `yaml:"defer,omitempty"`
}

// CloudInitConfig represents the cloud_init section of ~/.boite.yml
type CloudInitConfig struct {
	APT        *APTConfig        `yaml:"apt"`
	Packages   []string          `yaml:"packages"`
	Runcmd     []any             `yaml:"runcmd"`
	Users      []UserConfig      `yaml:"users"`
	WriteFiles []WriteFileConfig `yaml:"write_files"`
}

// APTConfig represents the apt sources configuration
type APTConfig struct {
	Sources map[string]APTSource `yaml:"sources"`
}

// APTSource represents a single apt source entry
type APTSource struct {
	KeyID  string `yaml:"keyid"`
	Source string `yaml:"source"`
}

// UserConfig represents a user definition in cloud-init
type UserConfig struct {
	Gecos             string   `yaml:"gecos"`
	Groups            []string `yaml:"groups"`
	Home              string   `yaml:"home"`
	Name              string   `yaml:"name"`
	Shell             string   `yaml:"shell"`
	Sudo              string   `yaml:"sudo"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Password          string   `yaml:"password,omitempty"`
}

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

// BoiteConfig represents the top-level ~/.boite.yml structure.
type BoiteConfig struct {
	CloudInit *CloudInitConfig `yaml:"cloud_init"`
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

// MergeCloudInitConfig merges the user's config with the defaults
func MergeCloudInitConfig(userCfg *BoiteConfig) *CloudInitConfig {
	if userCfg == nil || userCfg.CloudInit == nil {
		defaultCfg, err := defaultCloudInitConfig()
		if err != nil {
			return &CloudInitConfig{}
		}
		return defaultCfg
	}

	defaults, err := defaultCloudInitConfig()
	if err != nil {
		return userCfg.CloudInit
	}

	users := mergeUsers(defaults.Users, userCfg.CloudInit.Users)
	for i, user := range users {
		if user.Name == "boite" {
			users[i].SSHAuthorizedKeys = findAuthorizedKeys(userCfg, "boite")
		}
	}

	merged := &CloudInitConfig{
		APT:        userCfg.CloudInit.APT,
		Packages:   mergePackages(defaults.Packages, userCfg.CloudInit.Packages),
		Runcmd:     mergeRuncmd(defaults.Runcmd, userCfg.CloudInit.Runcmd),
		Users:      users,
		WriteFiles: mergeWriteFiles(defaults.WriteFiles, userCfg.CloudInit.WriteFiles),
	}

	return merged
}

func findAuthorizedKeys(cfg *BoiteConfig, name string) []string {
	if cfg == nil || cfg.CloudInit == nil {
		return nil
	}
	for _, u := range cfg.CloudInit.Users {
		if u.Name == name {
			return u.SSHAuthorizedKeys
		}
	}
	return nil
}

func mergePackages(defaults, user []string) []string {
	seen := make(map[string]bool, len(defaults))
	result := make([]string, 0, len(defaults)+len(user))
	for _, p := range defaults {
		seen[p] = true
		result = append(result, p)
	}
	for _, p := range user {
		if seen[p] {
			continue
		}
		seen[p] = true
		result = append(result, p)
	}
	return result
}

func mergeRuncmd(defaults, user []any) []any {
	result := make([]any, 0, len(defaults)+len(user))
	result = append(result, defaults...)
	result = append(result, user...)
	return result
}

func mergeUsers(defaults, user []UserConfig) []UserConfig {
	result := make([]UserConfig, 0, len(defaults))
	indexByName := make(map[string]int, len(defaults))
	for _, u := range defaults {
		indexByName[u.Name] = len(result)
		result = append(result, u)
	}
	for _, u := range user {
		if idx, ok := indexByName[u.Name]; ok {
			result[idx] = u
			continue
		}
		indexByName[u.Name] = len(result)
		result = append(result, u)
	}
	return result
}

func mergeWriteFiles(defaults, user []WriteFileConfig) []WriteFileConfig {
	result := make([]WriteFileConfig, 0, len(defaults))
	indexByPath := make(map[string]int, len(defaults))
	for _, w := range defaults {
		indexByPath[w.Path] = len(result)
		result = append(result, w)
	}
	for _, w := range user {
		if idx, ok := indexByPath[w.Path]; ok {
			result[idx] = w
			continue
		}
		indexByPath[w.Path] = len(result)
		result = append(result, w)
	}
	return result
}

// defaultCloudInitConfig returns the default cloud-init configuration by parsing the embedded YAML
func defaultCloudInitConfig() (*CloudInitConfig, error) {
	var cfg CloudInitConfig
	if err := yaml.Unmarshal([]byte(defaultCloudInitYAML), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse embedded cloudinit.yml: %w", err)
	}
	return &cfg, nil
}
