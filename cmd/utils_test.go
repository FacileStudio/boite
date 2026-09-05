package cmd

import (
	"os"
	"testing"
)

func TestHomeDir(t *testing.T) {
	home := homeDir()
	if home == "" {
		t.Fatal("expected non-empty home directory")
	}
}

func TestIsRoot(t *testing.T) {
	expected := os.Geteuid() == 0
	if isRoot() != expected {
		t.Fatalf("expected isRoot=%v, got %v", expected, isRoot())
	}
}

func TestCheckCommand(t *testing.T) {
	if err := checkCommand("nonexistentcommand12345"); err == nil {
		t.Fatal("expected error for nonexistent command, got nil")
	}
}

func TestDefaultCloudInit(t *testing.T) {
	cfg := defaultCloudInit()
	if len(cfg) == 0 {
		t.Fatal("expected non-empty cloud-init config")
	}
}

func TestVersionString(t *testing.T) {
	v := versionString()
	if v == "" {
		t.Fatal("expected non-empty version string")
	}
}

func TestEnsureCloudInit(t *testing.T) {
	tempDir := t.TempDir()
	path, err := ensureCloudInit(tempDir)
	if err != nil {
		t.Fatalf("unexpected error from ensureCloudInit: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated cloud-init: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected non-empty cloud-init file")
	}
}

func TestVMLaunchArgs(t *testing.T) {
	argsWithMount := vmLaunchArgs("test-box", "/path/to/ci.yml", "/workspace", false)
	if len(argsWithMount) == 0 {
		t.Fatal("expected non-empty launch args")
	}
	argsNoMount := vmLaunchArgs("test-box", "/path/to/ci.yml", "/workspace", true)
	if len(argsNoMount) >= len(argsWithMount) {
		t.Fatal("expected no-mount args to have fewer arguments than mount args")
	}
}
