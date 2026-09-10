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

	off := write("sync: false\n")
	if WorkspaceSyncEnabled(&Instance{}, off, false) {
		t.Fatal("expected sync: false to disable workspace sync")
	}

	on := write("sync: true\n")
	if !WorkspaceSyncEnabled(&Instance{}, on, false) {
		t.Fatal("expected sync: true to enable workspace sync")
	}

	unspecified := write("vm:\n  cpus: 1\n")
	if WorkspaceSyncEnabled(&Instance{}, unspecified, false) {
		t.Fatal("expected absent sync key to leave workspace sync disabled")
	}
}

func TestLoadBoiteConfigProvision(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "boite.yml")
	contents := "vm:\n  cpus: 2\nprovision:\n  packages:\n    - nala\n    - tmux\n  commands:\n    - 'echo hi'\n"
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBoiteConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("expected config to load")
	}
	if cfg.Provision == nil {
		t.Fatal("expected provision block to parse")
	}
	if len(cfg.Provision.Packages) != 2 || cfg.Provision.Packages[0] != "nala" || cfg.Provision.Packages[1] != "tmux" {
		t.Fatalf("unexpected packages: %v", cfg.Provision.Packages)
	}
	if len(cfg.Provision.Commands) != 1 || cfg.Provision.Commands[0] != "echo hi" {
		t.Fatalf("unexpected commands: %v", cfg.Provision.Commands)
	}
	if cfg.Provision.Empty() {
		t.Fatal("expected populated provision block to be non-empty")
	}
}

func TestProvisionEmpty(t *testing.T) {
	empty := &ProvisionConfig{}
	if !empty.Empty() {
		t.Fatal("expected empty provision block to report empty")
	}
}

func TestLoadBoiteConfigEnvLocal(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "boite.yml")
	contents := "env:\n  source: local\n  local:\n    vars:\n      LOG_LEVEL: debug\n    resolve:\n      - OPENROUTER_API_KEY\n"
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBoiteConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || cfg.Env == nil {
		t.Fatal("expected env block to parse")
	}
	if cfg.Env.EffectiveSource() != EnvSourceLocal {
		t.Fatalf("expected source local, got %q", cfg.Env.EffectiveSource())
	}
	if cfg.Env.Local.Vars["LOG_LEVEL"] != "debug" {
		t.Fatalf("expected LOG_LEVEL literal, got %q", cfg.Env.Local.Vars["LOG_LEVEL"])
	}
	if len(cfg.Env.Local.Resolve) != 1 || cfg.Env.Local.Resolve[0] != "OPENROUTER_API_KEY" {
		t.Fatalf("unexpected resolve list: %v", cfg.Env.Local.Resolve)
	}
	if cfg.Env.Casier.Project != "" {
		t.Fatalf("expected empty casier block, got %+v", cfg.Env.Casier)
	}
}

func TestLoadBoiteConfigEnvCasier(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "boite.yml")
	contents := "env:\n  source: casier\n  casier:\n    project: my-org\n    environment: dev\n    token_ref: CASIER_TOKEN\n"
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBoiteConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || cfg.Env == nil {
		t.Fatal("expected env block to parse")
	}
	if cfg.Env.EffectiveSource() != EnvSourceCasier {
		t.Fatalf("expected source casier, got %q", cfg.Env.EffectiveSource())
	}
	if cfg.Env.Casier.Project != "my-org" || cfg.Env.Casier.Environment != "dev" || cfg.Env.Casier.TokenRef != "CASIER_TOKEN" {
		t.Fatalf("unexpected casier block: %+v", cfg.Env.Casier)
	}
}

func TestEnvSourceDefaultsLocal(t *testing.T) {
	var e *EnvConfig
	if e.EffectiveSource() != EnvSourceLocal {
		t.Fatal("expected nil env block to default to local")
	}
	empty := &EnvConfig{}
	if empty.EffectiveSource() != EnvSourceLocal {
		t.Fatal("expected unset source to default to local")
	}
}
