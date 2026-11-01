package qemu

import (
	"net"
	"os"
	"strings"
)

// SSHBindAddrEnv names the environment variable that sets the host address
// QEMU binds each sandbox's forwarded SSH port to.
const SSHBindAddrEnv = "BOITE_SSH_BIND"

// SSHBindAddr returns the address the SSH hostfwd binds to: BOITE_SSH_BIND
// when it holds a valid IP, otherwise loopback so the CLI stays local-only.
func SSHBindAddr() string {
	bind := strings.TrimSpace(os.Getenv(SSHBindAddrEnv))
	if net.ParseIP(bind) == nil {
		return "127.0.0.1"
	}
	return bind
}

// SSHDialHost returns the host this machine uses to reach a sandbox's SSH
// port: the bind address, or loopback when the bind is a wildcard.
func SSHDialHost() string {
	bind := SSHBindAddr()
	if ip := net.ParseIP(bind); ip.IsUnspecified() {
		return "127.0.0.1"
	}
	return bind
}
