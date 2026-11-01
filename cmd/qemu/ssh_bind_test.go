package qemu

import (
	"strings"
	"testing"
)

func TestSSHBindAddr(t *testing.T) {
	cases := []struct {
		name, env, bind, dial string
	}{
		{"unset", "", "127.0.0.1", "127.0.0.1"},
		{"bridge", "172.18.0.1", "172.18.0.1", "172.18.0.1"},
		{"wildcard", "0.0.0.0", "0.0.0.0", "127.0.0.1"},
		{"garbage falls back to loopback", "lan", "127.0.0.1", "127.0.0.1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(SSHBindAddrEnv, c.env)
			if got := SSHBindAddr(); got != c.bind {
				t.Fatalf("bind = %q, want %q", got, c.bind)
			}
			if got := SSHDialHost(); got != c.dial {
				t.Fatalf("dial = %q, want %q", got, c.dial)
			}
		})
	}
}

func TestHostFwdUsesBindAddr(t *testing.T) {
	t.Setenv(SSHBindAddrEnv, "172.18.0.1")
	joined := strings.Join(BuildQEMUArgs(QEMUConfig{HostFwdPort: 2223}), " ")
	if !strings.Contains(joined, "hostfwd=tcp:172.18.0.1:2223-:22") {
		t.Fatalf("hostfwd not bound to BOITE_SSH_BIND: %s", joined)
	}
	t.Setenv(SSHBindAddrEnv, "")
	joined = strings.Join(BuildQEMUArgs(QEMUConfig{HostFwdPort: 2223}), " ")
	if !strings.Contains(joined, "hostfwd=tcp:127.0.0.1:2223-:22") {
		t.Fatalf("default hostfwd changed: %s", joined)
	}
}
