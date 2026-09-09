package qemu

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceSyncEnabled(t *testing.T) {
	inst := &Instance{}
	if WorkspaceSyncEnabled(inst, "", false) {
		t.Fatal("expected workspace sync disabled by default (opt-in required)")
	}
	if WorkspaceSyncEnabled(inst, "", true) {
		t.Fatal("expected forceSkip to disable workspace sync")
	}
	if WorkspaceSyncEnabled(&Instance{NoMount: true}, "", false) {
		t.Fatal("expected NoMount instance to disable workspace sync")
	}
	if WorkspaceSyncEnabled(&Instance{NoMount: true}, "", true) {
		t.Fatal("expected NoMount with forceSkip to disable workspace sync")
	}
}

func TestWorkspaceSyncEnabledConfigGate(t *testing.T) {
	dir := t.TempDir()

	write := func(contents string) string {
		p := filepath.Join(dir, "boite.yml")
		if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	off := write("workspace:\n  sync_at_run: false\n")
	if WorkspaceSyncEnabled(&Instance{}, off, false) {
		t.Fatal("expected sync_at_run: false to disable workspace sync")
	}

	on := write("workspace:\n  sync_at_run: true\n")
	if !WorkspaceSyncEnabled(&Instance{}, on, false) {
		t.Fatal("expected sync_at_run: true to enable workspace sync")
	}

	unspecified := write("vm:\n  cpus: 1\n")
	if WorkspaceSyncEnabled(&Instance{}, unspecified, false) {
		t.Fatal("expected absent workspace section to leave workspace sync disabled")
	}
}
