package qemu

import (
	"fmt"
	"strings"
)

// DefaultDistro is the default distribution when none is specified.
const DefaultDistro = "flux"

// LegacyDistro is the distribution of instances and flat package lists that
// predate the distro setting: an empty recorded distro means Debian.
const LegacyDistro = "debian"

// FluxImageSHA256 pins the Flux base image. Rebuild it with
// scripts/bake-flux.sh and update this together with the uploaded flux.qcow2.
const FluxImageSHA256 = "TODO-pin-flux-qcow2-sha256-after-ci-build"

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
	return []string{"flux", "debian", "alpine"}
}

// distroAliases maps retired distro names to the distro that replaced them.
// nixos instances and configs keep working as flux, the Facile NixOS build.
func distroAliases() map[string]string {
	return map[string]string{"nixos": "flux"}
}

func builtInDistros() map[string]DistroSpec {
	return map[string]DistroSpec{
		"flux": {
			Name:       "flux",
			ImageName:  "flux.qcow2",
			URL:        "https://boite.facile.studio/flux.qcow2",
			SHA256:     FluxImageSHA256,
			PkgManager: "nix",
		},
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
	}
}

// GetDistro returns the DistroSpec for the given name, defaulting to flux when
// empty and resolving aliases such as nixos.
func GetDistro(name string) (DistroSpec, error) {
	distro := strings.TrimSpace(strings.ToLower(name))
	if distro == "" {
		distro = DefaultDistro
	}
	if canonical, ok := distroAliases()[distro]; ok {
		distro = canonical
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
