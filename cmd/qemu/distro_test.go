package qemu

import (
	"strings"
	"testing"
)

func TestValidDistros(t *testing.T) {
	distros := ValidDistros()
	if len(distros) != 3 {
		t.Fatalf("expected 3 valid distros, got %d", len(distros))
	}
	expected := []string{"debian", "alpine", "nixos"}
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
		{name: "", wantName: "debian", wantPkg: "apt"},
		{name: "debian", wantName: "debian", wantPkg: "apt"},
		{name: "alpine", wantName: "alpine", wantPkg: "apk"},
		{name: "nixos", wantName: "nixos", wantPkg: "nix"},
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
	if !strings.HasSuffix(alpPath, "alpine.qcow2") {
		t.Errorf("expected alpine image path to end with alpine.qcow2, got: %s", alpPath)
	}

	nixPath := GetBaseImagePath("nixos")
	if !strings.HasSuffix(nixPath, "nixos.qcow2") {
		t.Errorf("expected nixos image path to end with nixos.qcow2, got: %s", nixPath)
	}
}
