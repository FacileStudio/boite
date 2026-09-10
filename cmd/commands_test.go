package cmd

import (
	"slices"
	"testing"
)

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
		"sync",
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

	noWorkspace := runCmd.Flags().Lookup("no-workspace")
	if noWorkspace == nil {
		t.Fatal("expected run command to have --no-workspace flag")
	}

	cfgFlag := rootCmd.PersistentFlags().Lookup("config")
	if cfgFlag == nil {
		t.Fatal("expected rootCmd to have persistent --config flag")
	}
}

func TestCommandAliases(t *testing.T) {
	var foundDestroy bool
	if slices.Contains(rmCmd.Aliases, "destroy") {
		foundDestroy = true
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
