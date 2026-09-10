package cmd

import (
	"os"
	"strings"
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

func TestBuildConfigDisk(t *testing.T) {
	if checkCommand("mkfs.fat") != nil && checkCommand("mcopy") != nil {
		t.Skip("mkfs.fat/mcopy not available")
	}
	dir := t.TempDir()
	diskPath, err := qemu.BuildConfigDisk(dir, "ssh-ed25519 test boite", nil)
	if err != nil {
		t.Fatalf("BuildConfigDisk: %v", err)
	}
	if _, err := os.Stat(diskPath); err != nil {
		t.Fatalf("config.img not created: %v", err)
	}
}

func TestBuildConfigDiskTiroirPayload(t *testing.T) {
	if checkCommand("mkfs.fat") != nil && checkCommand("mcopy") != nil {
		t.Skip("mkfs.fat/mcopy not available")
	}
	dir := t.TempDir()
	diskPath, err := qemu.BuildConfigDisk(dir, "ssh-ed25519 test boite", map[string]string{"LOG_LEVEL": "debug", "GREP_ME": "secret"})
	if err != nil {
		t.Fatalf("BuildConfigDisk: %v", err)
	}
	src, err := os.ReadFile(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), "secret") {
		t.Fatal("config disk must not contain the value in plaintext")
	}
}

func TestVersionString(t *testing.T) {
	v := versionString()
	if v == "" {
		t.Fatal("expected non-empty version string")
	}
}
