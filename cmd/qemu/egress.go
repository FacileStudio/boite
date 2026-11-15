package qemu

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FacileStudio/boite/cmd/egress"
)

// EgressProxyHost and EgressProxyPort are where an allowlist guest finds its
// proxy. The address sits inside slirp's guest subnet, outside the DHCP range.
const (
	EgressProxyHost = "10.0.2.100"
	EgressProxyPort = 3128
)

// EgressConfig locates the files of one allowlist VM: the extra hosts the
// proxy reads on every connection and the log that records denials. Exe is
// the boite binary QEMU re-runs as the proxy for each guest connection.
type EgressConfig struct {
	Exe       string
	HostsFile string
	LogFile   string
	Hosts     []string
}

// NewEgressConfig returns the egress files kept in an instance directory,
// pointing at the running binary.
func NewEgressConfig(instanceDir string, hosts []string) *EgressConfig {
	exe, _ := os.Executable()
	return &EgressConfig{
		Exe:       exe,
		HostsFile: filepath.Join(instanceDir, "egress-hosts.txt"),
		LogFile:   filepath.Join(instanceDir, "egress.log"),
		Hosts:     hosts,
	}
}

// Prepare writes the hosts file. A nil receiver is a no-op. It fails closed
// when the boite binary cannot be located.
func (e *EgressConfig) Prepare() error {
	if e == nil {
		return nil
	}
	if e.Exe == "" {
		return fmt.Errorf("allowlist net: cannot locate the boite binary that serves the egress proxy")
	}
	if err := egress.WriteHosts(e.HostsFile, e.Hosts); err != nil {
		return fmt.Errorf("write egress hosts: %w", err)
	}
	return nil
}

func (e *EgressConfig) guestfwd() string {
	cmd := strings.Join([]string{shellQuote(e.Exe), egress.HelperArg, shellQuote(e.HostsFile), shellQuote(e.LogFile)}, " ")
	cmd = strings.ReplaceAll(cmd, ",", ",,")
	return fmt.Sprintf("guestfwd=tcp:%s:%d-cmd:%s", EgressProxyHost, EgressProxyPort, cmd)
}
