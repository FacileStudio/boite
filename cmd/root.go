package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const asciiBanner = `▄▄
██          ▀▀  ██
████▄ ▄███▄ ██ ▀██▀▀ ▄█▀█▄
██ ██ ██ ██ ██  ██   ██▄█▀
████▀ ▀███▀ ██▄ ██   ▀█▄▄▄ `

var Version = "dev"
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
		return "v0.1.4"
	}
	return "v" + v
}

func init() {
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

	return ensureCloudInit(home)
}

func defaultCloudInit() string {
	return `#cloud-config
apt:
  sources:
    docker.list:
      source: https://download.docker.com/linux/ubuntu/gpg
      keyid: 9DC858229FC7DD38854AE2D88D81807C1AAF0A5B
packages:
  - docker-ce
  - docker-ce-cli
  - containerd.io
  - git
  - curl
  - unzip
  - ca-certificates
  - bash-completion
  - fzf
  - jq
  - tmux
  - wget
  - make
  - build-essential
  - zsh
  - zsh-antigen
  - neovim
users:
  - default
  - name: boite
    gecos: Boite
    groups: [sudo, docker]
    shell: /usr/bin/zsh
    sudo: ALL=(ALL) NOPASSWD:ALL
    home: /home/boite
runcmd:
  - chmod -x /etc/update-motd.d/* 2>/dev/null || true
  - rm -f /etc/legal /etc/motd
  - touch /home/boite/.hushlogin
  
  - |
    cat > /etc/motd << 'MOTD_EOF'
    ▄▄
    ██          ▀▀  ██
    ████▄ ▄███▄ ██ ▀██▀▀ ▄█▀█▄
    ██ ██ ██ ██ ██  ██   ██▄█▀
    ████▀ ▀███▀ ██▄ ██   ▀█▄▄▄

    v0.1.4
    MOTD_EOF
  - curl https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh
  - curl -L https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -
  - su - boite -c 'curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y'
  - su - boite -c 'curl -fsSL https://bun.sh/install | bash'
  - su - boite -c 'curl -fsSL https://raw.githubusercontent.com/saravenpi/skatos/main/install.sh | bash'
  - su - boite -c 'git config --global init.defaultBranch main'
  - su - boite -c 'git config --global safe.directory "*"'
  - >
    su - boite -c
    'sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" --unattended'
  - mkdir -p /workspace
  - chown -R boite:boite /workspace
  - |
    if [ -f /workspace/.zshrc_local ]; then
      cp /workspace/.zshrc_local /home/boite/.zshrc
      chown boite:boite /home/boite/.zshrc
    else
      cat > /home/boite/.zshrc << 'INNER_EOF'
export PATH="/usr/local/go/bin:/home/boite/.cargo/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.bun/bin:$PATH"
eval "$(/usr/local/bin/mise activate zsh)"
alias ll="ls -la"
alias ls="ls --color=auto"
alias grep="grep --color=auto"
INNER_EOF
      chown boite:boite /home/boite/.zshrc
    fi
  
  
  - mkdir -p /home/boite/.ssh
  
  - chown -R boite:boite /home/boite/.ssh
  - chmod 700 /home/boite/.ssh
  - test -f /home/boite/.ssh/authorized_keys && chmod 600 /home/boite/.ssh/authorized_keys
`
}

func createDefaultConfig(home string) error {
	defaultConfig := map[string]interface{}{
		"cloud_init": defaultCloudInit(),
		"vm": map[string]interface{}{
			"memory": "4G",
			"disk":   "40G",
			"cpus":   2,
		},
	}
	for k, v := range defaultConfig {
		viper.SetDefault(k, v)
	}
	configPath := filepath.Join(home, ".boite.yml")
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}
	fmt.Printf("Created default config at %s\n", configPath)
	return nil
}

func ensureCloudInit(home string) error {
	cloudInitStr := viper.GetString("cloud_init")
	if cloudInitStr == "" {
		legacyPath := filepath.Join(home, ".cloud-init.yml")
		b, err := os.ReadFile(legacyPath)
		if err != nil {
			return fmt.Errorf("cloud_init not set in config and no legacy .cloud-init.yml found: %w", err)
		}
		cloudInitStr = string(b)
	}

	cloudInitStr = strings.TrimLeft(cloudInitStr, "\n")
	if !strings.HasSuffix(cloudInitStr, "\n") {
		cloudInitStr += "\n"
	}

	dataDir := filepath.Join(home, "boite")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}
	cloudInitPath := filepath.Join(dataDir, "cloud-init.yml")
	if err := os.WriteFile(cloudInitPath, []byte(cloudInitStr), 0o644); err != nil {
		return fmt.Errorf("failed to write cloud-init to data file: %w", err)
	}
	viper.Set("cloud_init_path", cloudInitPath)
	return nil
}
