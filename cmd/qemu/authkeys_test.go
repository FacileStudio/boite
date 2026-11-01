package qemu

import "testing"

func TestJoinAuthorizedKeys(t *testing.T) {
	cases := []struct {
		name, control, primary string
		extra                  []string
		want                   string
	}{
		{"primary only", "", "k1", nil, "k1"},
		{"adds control", "ctl", "k1", nil, "k1\nctl"},
		{"user and control", "ctl", "k1", []string{"user"}, "k1\nuser\nctl"},
		{"dedupe and trim", "k1", " k1 ", []string{"", "k1"}, "k1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(ControlPubKeyEnv, c.control)
			if got := JoinAuthorizedKeys(c.primary, c.extra); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestValidatePublicKey(t *testing.T) {
	good := []string{"ssh-ed25519 AAAA user@host", "ecdsa-sha2-nistp256 AAAA", "sk-ssh-ed25519@openssh.com AAAA"}
	bad := []string{"", "ssh-ed25519", "ssh-ed25519 AAAA\nssh-rsa BBBB", "command=\"x\" ssh-ed25519 AAAA", "hello world"}
	for _, k := range good {
		if err := ValidatePublicKey(k); err != nil {
			t.Errorf("%q rejected: %v", k, err)
		}
	}
	for _, k := range bad {
		if ValidatePublicKey(k) == nil {
			t.Errorf("%q accepted", k)
		}
	}
}
