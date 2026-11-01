package worker

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	defaultSock = "/run/boite/worker.sock"
	sockMode    = 0o600
)

// Config is the worker's environment-derived settings.
type Config struct {
	Sock  string
	Token string
	Host  string
}

// LoadConfig reads BOITE_WORKER_SOCK, BOITE_WORKER_TOKEN, BOITE_WORKER_HOST
// and BOITE_SSH_BIND through getenv. It fails without a token or with an SSH
// bind that is not an IP address.
func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		Sock:  strings.TrimSpace(getenv("BOITE_WORKER_SOCK")),
		Token: getenv("BOITE_WORKER_TOKEN"),
		Host:  strings.TrimSpace(getenv("BOITE_WORKER_HOST")),
	}
	if cfg.Token == "" {
		return cfg, errors.New("BOITE_WORKER_TOKEN is required")
	}
	if cfg.Sock == "" {
		cfg.Sock = defaultSock
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
