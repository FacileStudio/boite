package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func TestDefaultCloudInitIsValidYAML(t *testing.T) {
	raw := defaultCloudInit()
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(raw), &node); err != nil {
		t.Fatalf("defaultCloudInit produced invalid YAML: %v", err)
	}

	var data struct {
		Runcmd []interface{} `yaml:"runcmd"`
		Users  []interface{} `yaml:"users"`
	}
	if err := yaml.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("failed to decode cloud-init fields: %v", err)
	}
	if len(data.Runcmd) == 0 {
		t.Fatal("expected non-empty runcmd")
	}
	if len(data.Users) == 0 {
		t.Fatal("expected non-empty users")
	}
}

func TestRootCommandStructure(t *testing.T) {
	if rootCmd.Use != "boite" {
		t.Fatalf("expected Use=boite, got %s", rootCmd.Use)
	}

	expectedSubcommands := []string{
		"create",
		"list",
		"shell",
		"start",
		"stop",
		"remove",
		"purge",
		"exec",
		"install",
	}

	commands := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		commands[c.Name()] = true
	}

	for _, name := range expectedSubcommands {
		if !commands[name] {
			t.Fatalf("expected subcommand '%s' to be registered", name)
		}
	}
}

func TestCommandFlags(t *testing.T) {
	flag := createCmd.Flags().Lookup("no-mount")
	if flag == nil {
		t.Fatal("expected create command to have --no-mount flag")
	}

	cfgFlag := rootCmd.PersistentFlags().Lookup("config")
	if cfgFlag == nil {
		t.Fatal("expected rootCmd to have persistent --config flag")
	}
}

func TestCommandAliases(t *testing.T) {
	var foundRm bool
	for _, alias := range rmCmd.Aliases {
		if alias == "rm" {
			foundRm = true
			break
		}
	}
	if !foundRm {
		t.Fatal("expected remove command to have 'rm' alias")
	}

	var foundSSH, foundEnter bool
	for _, alias := range shellCmd.Aliases {
		if alias == "ssh" {
			foundSSH = true
		}
		if alias == "enter" {
			foundEnter = true
		}
	}
	if !foundSSH || !foundEnter {
		t.Fatal("expected shell command to have 'ssh' and 'enter' aliases")
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Reset()
	if err := createDefaultConfig(tmpDir); err != nil {
		t.Fatalf("unexpected error creating default config: %v", err)
	}
	expectedPath := filepath.Join(tmpDir, ".boite.yml")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read created config: %v", err)
	}
	var parsed struct {
		VM        map[string]interface{} `yaml:"vm"`
		CloudInit map[string]interface{} `yaml:"cloud_init"`
	}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("written config contains invalid YAML: %v", err)
	}
	if len(parsed.VM) == 0 {
		t.Fatal("expected vm settings in default config")
	}
	if len(parsed.CloudInit) == 0 {
		t.Fatal("expected cloud_init mapping in default config")
	}
}

func TestEnsureCloudInitWithStructuredYAML(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Reset()
	viper.Set("cloud_init", map[string]interface{}{
		"packages": []string{"git", "curl"},
		"runcmd":   []string{"echo hello"},
	})
	path, err := ensureCloudInit(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error with structured YAML: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated cloud-init: %v", err)
	}
	if !strings.HasPrefix(string(content), "#cloud-config\n") {
		t.Fatalf("expected #cloud-config header, got: %s", string(content))
	}
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		t.Fatalf("generated cloud-init is not valid YAML: %v", err)
	}
}

func TestEnsureCloudInitWithCorruptedConfig(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Reset()
	viper.Set("cloud_init", "invalid: [yaml: broken: {{{")
	path, err := ensureCloudInit(tmpDir)
	if err != nil {
		t.Fatalf("expected ensureCloudInit to heal corrupted config, got error: %v", err)
	}
	if path == "" {
		t.Fatal("expected valid cloud-init path returned")
	}
}
