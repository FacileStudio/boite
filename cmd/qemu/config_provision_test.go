package qemu

import (
	"os"
	"path/filepath"
	"testing"
)

func verifyDistroPackages(t *testing.T, cfg *BoiteConfig, distro, expectedSpecific string) {
	pkgs, _ := cfg.Provision.ForDistro(distro)
	if len(pkgs) != 2 || pkgs[0] != "curl" || pkgs[1] != expectedSpecific {
		t.Fatalf("unexpected %s packages: %v", distro, pkgs)
	}
}

func TestDistroSpecificPackages(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "boite.yml")
	contents := "provision:\n  packages:\n    common:\n      - curl\n    debian:\n      - nala\n    alpine:\n      - bash\n    nixos:\n      - ripgrep\n  commands:\n    - echo hi\n"
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBoiteConfig(p)
	if err != nil || cfg == nil || cfg.Provision == nil {
		t.Fatal("expected provision block to parse")
	}

	verifyDistroPackages(t, cfg, "debian", "nala")
	verifyDistroPackages(t, cfg, "alpine", "bash")
	verifyDistroPackages(t, cfg, "nixos", "ripgrep")
}

func TestFlatListPackagesOnlyAppliesToDefaultDistro(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "boite.yml")
	contents := "provision:\n  packages:\n    - nala\n"
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBoiteConfig(p)
	if err != nil {
		t.Fatal(err)
	}

	debPkgs, _ := cfg.Provision.ForDistro("debian")
	if len(debPkgs) != 1 || debPkgs[0] != "nala" {
		t.Fatalf("expected nala for debian, got %v", debPkgs)
	}

	nixPkgs, _ := cfg.Provision.ForDistro("nixos")
	if len(nixPkgs) != 0 {
		t.Fatalf("expected no packages for nixos when flat list, got %v", nixPkgs)
	}
}
