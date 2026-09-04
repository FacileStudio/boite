package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set during build
var Version = "dev"

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "boite",
	Short: "Dev Sandbox Manager",
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
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.boite.yml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	var home string
	var err error
	
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not find home directory: %w", err)
		}

		// Search config in home directory with name ".boite" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yml")
		viper.SetConfigName(".boite")
	}

	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; create a default one
		defaultCloudInit := `#cloud-config
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
runcmd:
  # Install mise
  - curl https://mise.run | sh

  # Install Go
  - curl -L https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -

  # Install Rust via rustup
  - curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y

  # Install Bun
  - npm install -g bun@1.3.14

  # Install skatos
  - curl -fsSL https://raw.githubusercontent.com/saravenpi/skatos/main/install.sh | bash

  # Configure git
  - git config --global init.defaultBranch main
  - git config --global safe.directory '*'

  # Install and set zsh as default shell
  - chsh -s /usr/bin/zsh ubuntu
  - su - ubuntu -c 'sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" --unattended'

  # Set up PATH for user
  - mkdir -p /workspace
  - chown ubuntu:ubuntu /workspace
  - |
    if [ -f /workspace/.zshrc_local ]; then
      cp /workspace/.zshrc_local /home/ubuntu/.zshrc
      chown ubuntu:ubuntu /home/ubuntu/.zshrc
    else
      cat > /home/ubuntu/.zshrc << 'EOF'
export PATH="/usr/local/go/bin:/root/.cargo/bin:/root/.local/bin:$HOME/.bun/bin:$PATH"
eval "$(/root/.local/bin/mise activate zsh)"
alias ll="ls -la"
alias ls="ls --color=auto"
alias grep="grep --color=auto"
EOF
      chown ubuntu:ubuntu /home/ubuntu/.zshrc
    fi

  # Enable docker for ubuntu user (without sudo)
  - usermod -aG docker ubuntu
  - "test -f /workspace/.zshrc_local && cat /workspace/.zshrc_local >> /home/ubuntu/.zshrc\"
		defaultConfig := map[string]interface{}{
			"cloud_init": defaultCloudInit,
			"vm": map[string]interface{}{
				"memory": "4G",
				"disk":   "40G",
				"cpus":   2,
			},
		}
		for k, v := range defaultConfig {
			viper.SetDefault(k, v)
		}
		// Write the default config
		configPath := filepath.Join(home, ".boite.yml")
		if err := viper.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write default config: %w", err)
		}
		fmt.Printf("Created default config at %s\n", configPath)
	}

	// Ensure we have a cloud-init string to pass to multipass
	cloudInitStr := viper.GetString("cloud_init")
	if cloudInitStr == "" {
		// fallback to legacy path if user upgraded
		legacyPath := filepath.Join(home, ".cloud-init.yml")
		if _, err := os.Stat(legacyPath); err == nil {
			b, err := os.ReadFile(legacyPath)
			if err != nil {
				return fmt.Errorf("failed to read legacy cloud-init file: %w", err)
			}
			cloudInitStr = string(b)
		} else {
			return fmt.Errorf("cloud_init not set in config and no legacy .cloud-init.yml found")
		}
	}
	// Remove possible leading newline from block scalar
	cloudInitStr = strings.TrimLeft(cloudInitStr, "\n")
	// Ensure it ends with newline
	if !strings.HasSuffix(cloudInitStr, "\n") {
		cloudInitStr += "\n"
	}
	// Write cloud-init to a fixed file in user's boite dir (accessible to snap, non-hidden)
	dataDir := filepath.Join(homeDir(), "boite")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}
	cloudInitPath := filepath.Join(dataDir, "cloud-init.yml")
	if err := os.WriteFile(cloudInitPath, []byte(cloudInitStr), 0o644); err != nil {
		return fmt.Errorf("failed to write cloud-init to data file: %w", err)
	}
	// Override config with the data file path so subprocesses see it
	viper.Set("cloud_init_path", cloudInitPath)

	return nil
}