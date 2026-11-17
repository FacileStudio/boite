package qemu

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidDistros(t *testing.T) {
	distros := ValidDistros()
	if len(distros) != 3 {
		t.Fatalf("expected 3 valid distros, got %d", len(distros))
	}
	expected := []string{"flux", "debian", "alpine"}
	for i, name := range expected {
		if distros[i] != name {
			t.Errorf("expected distro %d to be %s, got %s", i, name, distros[i])
		}
	}
}

func TestGetDistro(t *testing.T) {
	tests := []struct {
		name       string
		wantName   string
		wantPkg    string
		shouldFail bool
	}{
		{name: "", wantName: "flux", wantPkg: "nix"},
		{name: "flux", wantName: "flux", wantPkg: "nix"},
		{name: "debian", wantName: "debian", wantPkg: "apt"},
		{name: "alpine", wantName: "alpine", wantPkg: "apk"},
		{name: "nixos", wantName: "flux", wantPkg: "nix"},
		{name: " NixOS ", wantName: "flux", wantPkg: "nix"},
		{name: "ubuntu", shouldFail: true},
	}

	for _, tt := range tests {
		spec, err := GetDistro(tt.name)
		if tt.shouldFail {
			if err == nil {
				t.Errorf("GetDistro(%q) expected error, got nil", tt.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("GetDistro(%q) unexpected error: %v", tt.name, err)
			continue
		}
		if spec.Name != tt.wantName || spec.PkgManager != tt.wantPkg {
			t.Errorf("GetDistro(%q) = (%s, %s), want (%s, %s)", tt.name, spec.Name, spec.PkgManager, tt.wantPkg, tt.wantName)
		}
	}
}

func TestEffectiveDistro(t *testing.T) {
	var nilCfg *VMConfig
	if nilCfg.EffectiveDistro() != DefaultDistro {
		t.Fatalf("expected nil config to return default distro")
	}

	emptyCfg := &VMConfig{}
	if emptyCfg.EffectiveDistro() != DefaultDistro {
		t.Fatalf("expected empty config to return default distro")
	}

	customCfg := &VMConfig{Distro: "alpine"}
	if customCfg.EffectiveDistro() != "alpine" {
		t.Fatalf("expected custom config to return alpine")
	}
}

func TestGetBaseImagePath(t *testing.T) {
	alpPath := GetBaseImagePath("alpine")
	if !strings.Contains(filepath.Base(alpPath), "alpine") || !strings.HasSuffix(alpPath, ".qcow2") {
		t.Errorf("expected an alpine qcow2 image path, got: %s", alpPath)
	}

	want := "flux-" + FluxImageSHA256[:12] + ".qcow2"
	for _, name := range []string{"", "flux", "nixos"} {
		if p := GetBaseImagePath(name); filepath.Base(p) != want {
			t.Errorf("expected %q image path to end with %s, got: %s", name, want, p)
		}
	}
}
