package cmd

import (
	"os"
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

func TestBuildConfigISO(t *testing.T) {
	if err := checkCommand("genisoimage"); err != nil {
		t.Skip("genisoimage not available")
	}
	dir := t.TempDir()
	isoPath, err := qemu.BuildConfigISO(dir, "ssh-ed25519 test boite")
	if err != nil {
		t.Fatalf("BuildConfigISO: %v", err)
	}
	if _, err := os.Stat(isoPath); err != nil {
		t.Fatalf("config.iso not created: %v", err)
	}
}

func TestVersionString(t *testing.T) {
	v := versionString()
	if v == "" {
		t.Fatal("expected non-empty version string")
	}
}
