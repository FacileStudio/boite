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
