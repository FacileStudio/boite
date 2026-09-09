package cmd

import (
	"testing"

	"github.com/FacileStudio/boite/cmd/qemu"
)

func TestHomeDir(t *testing.T) {
	home := homeDir()
	if home == "" {
		t.Fatal("expected non-empty home directory")
	}
}

func TestCheckCommand(t *testing.T) {
	if err := checkCommand("nonexistentcommand12345"); err == nil {
		t.Fatal("expected error for nonexistent command, got nil")
	}
}

func TestDefaultCloudInit(t *testing.T) {
	cfg := qemu.DefaultCloudInitYAML()
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
