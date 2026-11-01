package worker

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
)

const (
	defaultSock     = "/run/boite/worker.sock"
	defaultSockMode = os.FileMode(0o600)
	noGroup         = -1
)

// Config is the worker's environment-derived settings.
type Config struct {
	Sock     string
	Token    string
	Host     string
	SockMode os.FileMode
	SockGID  int
}

// LoadConfig reads BOITE_WORKER_SOCK, BOITE_WORKER_TOKEN, BOITE_WORKER_HOST,
// BOITE_WORKER_SOCK_MODE, BOITE_WORKER_SOCK_GROUP and BOITE_SSH_BIND through
// getenv. It fails without a token, with an SSH bind that is not an IP
// address, or with a socket mode or group it cannot accept.
func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		Sock:    strings.TrimSpace(getenv("BOITE_WORKER_SOCK")),
		Token:   getenv("BOITE_WORKER_TOKEN"),
		Host:    strings.TrimSpace(getenv("BOITE_WORKER_HOST")),
		SockGID: noGroup,
	}
	if cfg.Token == "" {
		return cfg, errors.New("BOITE_WORKER_TOKEN is required")
	}
	if cfg.Sock == "" {
		cfg.Sock = defaultSock
	}
	if err := cfg.loadSocketAccess(getenv); err != nil {
		return cfg, err
	}
	bind := strings.TrimSpace(getenv("BOITE_SSH_BIND"))
	if bind != "" && net.ParseIP(bind) == nil {
		return cfg, fmt.Errorf("BOITE_SSH_BIND %q is not an IP address", bind)
	}
	if cfg.Host == "" {
		cfg.Host = bind
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	return cfg, nil
}

func (c *Config) loadSocketAccess(getenv func(string) string) error {
	mode, err := parseSockMode(strings.TrimSpace(getenv("BOITE_WORKER_SOCK_MODE")))
	if err != nil {
		return err
	}
	c.SockMode = mode
	group := strings.TrimSpace(getenv("BOITE_WORKER_SOCK_GROUP"))
	if group == "" {
		return nil
	}
	gid, err := resolveGroup(group)
	if err != nil {
		return err
	}
	if mode&0o060 != 0o060 {
		return fmt.Errorf("BOITE_WORKER_SOCK_GROUP needs BOITE_WORKER_SOCK_MODE with group read and write, such as 0660; got %04o", mode)
	}
	c.SockGID = gid
	return nil
}

func parseSockMode(raw string) (os.FileMode, error) {
	if raw == "" {
		return defaultSockMode, nil
	}
	n, err := strconv.ParseUint(raw, 8, 32)
	if err != nil || n > 0o777 {
		return 0, fmt.Errorf("BOITE_WORKER_SOCK_MODE %q is not an octal mode up to 0777", raw)
	}
	mode := os.FileMode(n)
	if mode&0o600 != 0o600 {
		return 0, fmt.Errorf("BOITE_WORKER_SOCK_MODE %q must give the owner read and write", raw)
	}
	if mode&0o007 != 0 {
		return 0, fmt.Errorf("BOITE_WORKER_SOCK_MODE %q must not open the socket to other users; use BOITE_WORKER_SOCK_GROUP", raw)
	}
	return mode, nil
}

func resolveGroup(raw string) (int, error) {
	if gid, err := strconv.Atoi(raw); err == nil {
		if gid < 0 {
			return 0, fmt.Errorf("BOITE_WORKER_SOCK_GROUP %q is not a valid gid", raw)
		}
		return gid, nil
	}
	grp, err := user.LookupGroup(raw)
	if err != nil {
		return 0, fmt.Errorf("BOITE_WORKER_SOCK_GROUP %q: %w", raw, err)
	}
	gid, err := strconv.Atoi(grp.Gid)
	if err != nil {
		return 0, fmt.Errorf("BOITE_WORKER_SOCK_GROUP %q has a non-numeric gid %q", raw, grp.Gid)
	}
	return gid, nil
}
