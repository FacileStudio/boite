package cmd

import (
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

// findCommand returns the subcommand registered under name, or nil.
func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, c := range root.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRootCommandStructure(t *testing.T) {
	rootCmd := newRootCmd()
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
		"env",
	}

	for _, name := range expectedSubcommands {
		if findCommand(rootCmd, name) == nil {
			t.Fatalf("expected subcommand '%s' to be registered", name)
		}
	}
}

func TestCommandFlags(t *testing.T) {
	rootCmd := newRootCmd()

	createCmd := findCommand(rootCmd, "create")
	if createCmd == nil {
		t.Fatal("expected create command to exist")
	}
	flag := createCmd.Flags().Lookup("no-mount")
	if flag == nil {
		t.Fatal("expected create command to have --no-mount flag")
	}

	runCmd := findCommand(rootCmd, "run")
	if runCmd == nil {
		t.Fatal("expected run command to exist")
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
	rootCmd := newRootCmd()
	rmCmd := findCommand(rootCmd, "rm")
	runCmd := findCommand(rootCmd, "run")
	if rmCmd == nil || runCmd == nil {
		t.Fatal("expected rm and run commands to exist")
	}
	if !slices.Contains(rmCmd.Aliases, "destroy") {
		t.Fatal("expected rm command to have 'destroy' alias")
	}
	for _, alias := range []string{"shell", "ssh", "enter"} {
		if !slices.Contains(runCmd.Aliases, alias) {
			t.Fatalf("expected run command to have '%s' alias", alias)
		}
	}
}
