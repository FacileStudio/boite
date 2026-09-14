package qemu

import (
	"strings"
	"testing"
)

func TestEffectiveNet(t *testing.T) {
	cases := []struct {
		net  string
		want string
		err  bool
	}{
		{"", "open", false},
		{"open", "open", false},
		{"offline", "offline", false},
		{"transparent", "", true},
	}
	for _, c := range cases {
		got, err := (&VMConfig{Net: c.net}).EffectiveNet()
		if (err != nil) != c.err {
			t.Fatalf("net %q: err %v, wantErr %v", c.net, err, c.err)
		}
		if !c.err && got != c.want {
			t.Fatalf("net %q: expected %q, got %q", c.net, c.want, got)
		}
	}
	got, err := (*VMConfig)(nil).EffectiveNet()
	if err != nil || got != "open" {
		t.Fatalf("nil VMConfig: expected open, got %q err %v", got, err)
	}
}

func TestBuildQEMUArgsNetRestrict(t *testing.T) {
	open := BuildQEMUArgs(QEMUConfig{HostFwdPort: 2222})
	if !containsArg(open, "restrict=off") {
		t.Fatal("expected restrict=off by default")
	}
	closed := BuildQEMUArgs(QEMUConfig{HostFwdPort: 2222, RestrictNet: true})
	if !containsArg(closed, "restrict=on") {
		t.Fatal("expected restrict=on when RestrictNet set")
	}
	if !containsArg(closed, "hostfwd=tcp:127.0.0.1:2222-:22") {
		t.Fatal("expected SSH hostfwd to survive restrict=on")
	}
}

func containsArg(args []string, substr string) bool {
	for _, a := range args {
		if strings.Contains(a, substr) {
			return true
		}
	}
	return false
}
