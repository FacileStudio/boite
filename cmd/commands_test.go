package cmd

import (
	"testing"

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
		"run",
		"start",
		"stop",
		"rm",
		"exec",
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
	var foundDestroy bool
	for _, alias := range rmCmd.Aliases {
		if alias == "destroy" {
			foundDestroy = true
			break
		}
	}
	if !foundDestroy {
		t.Fatal("expected rm command to have 'destroy' alias")
	}

	var foundShell, foundSSH, foundEnter bool
	for _, alias := range runCmd.Aliases {
		if alias == "shell" {
			foundShell = true
		}
		if alias == "ssh" {
			foundSSH = true
		}
		if alias == "enter" {
			foundEnter = true
		}
	}
	if !foundShell || !foundSSH || !foundEnter {
		t.Fatal("expected run command to have 'shell', 'ssh', and 'enter' aliases")
	}
}
