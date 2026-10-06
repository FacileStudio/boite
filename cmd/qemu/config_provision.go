package qemu

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type PackageList []string

func (p *PackageList) UnmarshalYAML(value *yaml.Node) error {
	var single string
	if err := value.Decode(&single); err == nil {
		*p = []string{single}
		return nil
	}
	var list []string
	if err := value.Decode(&list); err != nil {
		return err
	}
	*p = list
	return nil
}

type PackagesConfig struct {
	Common  PackageList            `yaml:"common"`
	Debian  PackageList            `yaml:"debian"`
	Alpine  PackageList            `yaml:"alpine"`
	NixOS   PackageList            `yaml:"nixos"`
	Distros map[string]PackageList `yaml:"-"`
	List    PackageList            `yaml:"-"`
}

func (p *PackagesConfig) UnmarshalYAML(value *yaml.Node) error {
	var list PackageList
	if err := value.Decode(&list); err == nil {
		p.List = list
		return nil
	}
	type rawPackages PackagesConfig
	var raw rawPackages
	if err := value.Decode(&raw); err != nil {
		return err
	}
	*p = PackagesConfig(raw)
	var dynamic map[string]PackageList
	if err := value.Decode(&dynamic); err == nil {
		p.Distros = dynamic
	}
	return nil
}

func (p *PackagesConfig) specificFor(name string) []string {
	switch name {
	case "debian":
		return p.Debian
	case "alpine":
		return p.Alpine
	case "nixos":
		return p.NixOS
	default:
		if p.Distros != nil {
			return p.Distros[name]
		}
		return nil
	}
}

func (p *PackagesConfig) ForDistro(distro string) []string {
	if p == nil {
		return nil
	}
	if len(p.List) > 0 {
		if strings.EqualFold(distro, DefaultDistro) {
			return p.List
		}
		return nil
	}
	res := append([]string{}, p.Common...)
	specific := p.specificFor(strings.ToLower(distro))
	return append(res, specific...)
}

type ProvisionConfig struct {
	Packages PackagesConfig `yaml:"packages"`
	Commands []string       `yaml:"commands"`
}

func (p *ProvisionConfig) ForDistro(distro string) ([]string, []string) {
	if p == nil {
		return nil, nil
	}
	return p.Packages.ForDistro(distro), p.Commands
}

func (p *ProvisionConfig) Empty() bool {
	if p == nil {
		return true
	}
	hasPkgs := len(p.Packages.List) > 0 || len(p.Packages.Common) > 0 ||
		len(p.Packages.Debian) > 0 || len(p.Packages.Alpine) > 0 ||
		len(p.Packages.NixOS) > 0 || len(p.Packages.Distros) > 0
	return !hasPkgs && len(p.Commands) == 0
}
