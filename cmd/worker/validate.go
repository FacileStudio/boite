package worker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/FacileStudio/boite/cmd/qemu"
)

const (
	defaultCPUs   = 2
	maxCPUs       = 32
	defaultMemory = "2G"
	defaultDisk   = "20G"
	maxOwnerLen   = 128
)

var (
	namePattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)
	memoryPattern = regexp.MustCompile(`^([1-9][0-9]{0,5})([MG])$`)
	diskPattern   = regexp.MustCompile(`^([1-9][0-9]{0,3})G?$`)
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// ValidID reports whether id is a safe box name: lowercase letters, digits
// and dashes, at most 40 characters, starting with a letter or digit.
func ValidID(id string) bool {
	return namePattern.MatchString(id)
}

// Normalize checks a create request and returns it with defaults applied.
func (r CreateRequest) Normalize() (CreateRequest, error) {
	if !ValidID(r.Name) {
		return r, invalid("name must match %s", namePattern)
	}
	owner := strings.TrimSpace(r.Owner)
	if owner == "" || len(owner) > maxOwnerLen || strings.ContainsAny(owner, "\r\n\x00") {
		return r, invalid("owner is required, one line, at most %d characters", maxOwnerLen)
	}
	r.Owner = owner
	if err := r.normalizeResources(); err != nil {
		return r, err
	}
	net, err := NormalizeNet(r.Net)
	if err != nil {
		return r, err
	}
	r.Net = net
	if r.SSHKey != "" {
		if err := qemu.ValidatePublicKey(r.SSHKey); err != nil {
			return r, invalid("ssh_key: %v", err)
		}
		r.SSHKey = strings.TrimSpace(r.SSHKey)
	}
	return r, nil
}

func (r *CreateRequest) normalizeResources() error {
	if r.CPUs == 0 {
		r.CPUs = defaultCPUs
	}
	if r.CPUs < 1 || r.CPUs > maxCPUs {
		return invalid("cpus must be between 1 and %d", maxCPUs)
	}
	r.Memory = strings.ToUpper(strings.TrimSpace(r.Memory))
	if r.Memory == "" {
		r.Memory = defaultMemory
	}
	if !memoryPattern.MatchString(r.Memory) {
		return invalid("memory must look like 512M or 2G")
	}
	r.Disk = strings.ToUpper(strings.TrimSpace(r.Disk))
	if r.Disk == "" {
		r.Disk = defaultDisk
	}
	match := diskPattern.FindStringSubmatch(r.Disk)
	if match == nil {
		return invalid("disk must look like 20G")
	}
	r.Disk = match[1] + "G"
	return nil
}

// NormalizeNet returns the egress mode net names, NetOpen when it is empty,
// or ErrInvalid when it is neither open nor offline.
func NormalizeNet(net string) (string, error) {
	mode, err := (&qemu.VMConfig{Net: net}).EffectiveNet()
	if err != nil {
		return "", invalid("net must be %s or %s", NetOpen, NetOffline)
	}
	return mode, nil
}
