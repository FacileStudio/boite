package qemu

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CloudInitConfig represents the cloud_init section of ~/.boite.yml
type CloudInitConfig struct {
	APT      *APTConfig     `yaml:"apt"`
	Packages []string       `yaml:"packages"`
	Runcmd   []interface{}  `yaml:"runcmd"`
	Users    []UserConfig   `yaml:"users"`
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
	Gecos            string   `yaml:"gecos"`
	Groups           []string `yaml:"groups"`
	Home             string   `yaml:"home"`
	Name             string   `yaml:"name"`
	Shell            string   `yaml:"shell"`
	Sudo             string   `yaml:"sudo"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

// VMConfig represents the vm section of ~/.boite.yml
type VMConfig struct {
	CPUs   int    `yaml:"cpus"`
	Disk   string `yaml:"disk"`
	Memory string `yaml:"memory"`
}

// BoiteConfig represents the full ~/.boite.yml structure
type BoiteConfig struct {
	CloudInit *CloudInitConfig `yaml:"cloud_init"`
	VM        *VMConfig        `yaml:"vm"`
}

// LoadBoiteConfig reads and parses ~/.boite.yml, returning a BoiteConfig
func LoadBoiteConfig() (*BoiteConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}

	configPath := filepath.Join(home, ".boite.yml")
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
	defaults := defaultCloudInitConfig()

	if userCfg == nil || userCfg.CloudInit == nil {
		return defaults
	}

	// Start fresh: user config overrides defaults, no duplication
	merged := &CloudInitConfig{
		Packages:  []string{},
		Runcmd:    []interface{}{},
		Users:     []UserConfig{},
		APT:       userCfg.CloudInit.APT,
	}

	// Packages: defaults first, then user adds new ones (deduplicated)
	seenPkgs := make(map[string]bool)
	for _, p := range defaults.Packages {
		seenPkgs[p] = true
		merged.Packages = append(merged.Packages, p)
	}
	for _, p := range userCfg.CloudInit.Packages {
		if !seenPkgs[p] {
			seenPkgs[p] = true
			merged.Packages = append(merged.Packages, p)
		}
	}

	// Runcmd: defaults first, then user commands (no dedup since order matters)
	merged.Runcmd = append(merged.Runcmd, defaults.Runcmd...)
	merged.Runcmd = append(merged.Runcmd, userCfg.CloudInit.Runcmd...)

	// Users: merge by name, user config overrides default if same name
	seenUsers := make(map[string]bool)
	for _, u := range defaults.Users {
		seenUsers[u.Name] = true
		merged.Users = append(merged.Users, u)
	}
	for _, u := range userCfg.CloudInit.Users {
		if seenUsers[u.Name] {
			// Replace existing user
			for i, mu := range merged.Users {
				if mu.Name == u.Name {
					merged.Users[i] = u
					break
				}
			}
		} else {
			seenUsers[u.Name] = true
			merged.Users = append(merged.Users, u)
		}
	}

	return merged
}

// defaultCloudInitConfig returns the default cloud-init configuration
func defaultCloudInitConfig() *CloudInitConfig {
	return &CloudInitConfig{
		Packages: []string{
			"git",
			"curl",
			"unzip",
			"ca-certificates",
			"bash-completion",
			"fzf",
			"jq",
			"tmux",
			"wget",
			"make",
			"build-essential",
			"zsh",
			"neovim",
			"qemu-guest-agent",
		},
		Runcmd: []interface{}{
			"rm -f /etc/legal /etc/motd",
			"touch /home/boite/.hushlogin",
			"curl https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh",
			"curl -L https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -",
			"su - boite -c 'curl --proto \"=https\" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y'",
			"su - boite -c 'curl -fsSL https://bun.sh/install | bash'",
			"su - boite -c 'curl -fsSL https://raw.githubusercontent.com/saravenpi/skatos/main/install.sh | bash'",
			"su - boite -c 'git config --global init.defaultBranch main'",
			"su - boite -c 'git config --global safe.directory \"*\"'",
			"su - boite -c 'sh -c \"$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)\" --unattended'",
			"mkdir -p /workspace",
			"chown -R boite:boite /workspace",
			"mkdir -p /home/boite/.ssh",
			"chown -R boite:boite /home/boite/.ssh",
			"chmod 700 /home/boite/.ssh",
			"test -f /home/boite/.ssh/authorized_keys && chmod 600 /home/boite/.ssh/authorized_keys",
		},
		Users: []UserConfig{
			{
				Gecos:  "Boite",
				Groups: []string{"sudo"},
				Home:   "/home/boite",
				Name:   "boite",
				Shell:  "/bin/bash",
				Sudo:   "ALL=(ALL) NOPASSWD:ALL",
				SSHAuthorizedKeys: []string{},
			},
		},
	}
}
