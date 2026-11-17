package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCreateDistroCanonicalizesNixos(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "boite.yml")
	if err := os.WriteFile(cfgPath, []byte("vm:\n  distro: nixos\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing.yml")
	cases := []struct{ flag, cfg, want string }{
		{"", cfgPath, "flux"},
		{"nixos", missing, "flux"},
		{"", missing, "flux"},
		{"debian", cfgPath, "debian"},
	}
	for _, c := range cases {
		got, err := resolveCreateDistro(c.flag, c.cfg)
		if err != nil || got != c.want {
			t.Errorf("resolveCreateDistro(%q, %q) = %q, %v; want %q", c.flag, c.cfg, got, err, c.want)
		}
	}
}
