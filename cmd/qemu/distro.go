package qemu

import (
	"fmt"
	"strings"
)

// DefaultDistro is the default distribution when none is specified.
const DefaultDistro = "debian"

// DistroSpec defines the image and package configuration for a guest distribution.
type DistroSpec struct {
	Name       string
	ImageName  string
	URL        string
	SHA256     string
	PkgManager string
}

// ValidDistros returns the list of supported distribution names.
func ValidDistros() []string {
	return []string{"debian", "alpine", "nixos"}
}

func builtInDistros() map[string]DistroSpec {
	return map[string]DistroSpec{
		"debian": {
			Name:       "debian",
			ImageName:  "debian.qcow2",
			URL:        BaseImageURL,
			SHA256:     BaseImageSHA256,
			PkgManager: "apt",
		},
		"alpine": {
			Name:       "alpine",
			ImageName:  "alpine.qcow2",
			URL:        "https://boite.facile.studio/alpine.qcow2",
			SHA256:     "",
			PkgManager: "apk",
		},
		"nixos": {
			Name:       "nixos",
			ImageName:  "nixos.qcow2",
			URL:        "https://boite.facile.studio/nixos.qcow2",
			SHA256:     "",
			PkgManager: "nix",
		},
	}
}

// GetDistro returns the DistroSpec for the given name, defaulting to debian when empty.
func GetDistro(name string) (DistroSpec, error) {
	distro := strings.TrimSpace(strings.ToLower(name))
	if distro == "" {
		distro = DefaultDistro
	}
	spec, ok := builtInDistros()[distro]
	if !ok {
		return DistroSpec{}, fmt.Errorf("unsupported distro %q (valid: %s)", name, strings.Join(ValidDistros(), ", "))
	}
	return spec, nil
}

// ResolveDistro resolves the distro from an explicit name or config fallback.
func ResolveDistro(name string, cfg *BoiteConfig) string {
	if name != "" {
		return name
	}
	if cfg != nil {
		return cfg.VM.EffectiveDistro()
	}
	return DefaultDistro
}
