package qemu

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEffectiveNetAllowlist(t *testing.T) {
	v := &VMConfig{Net: " Allowlist ", NetAllow: []string{"B.com", "a.com", "b.com"}}
	mode, err := v.EffectiveNet()
	if err != nil || mode != NetAllowlist {
		t.Fatalf("mode = %q, %v", mode, err)
	}
	hosts, err := v.EffectiveNetAllow()
	if err != nil || strings.Join(hosts, ",") != "b.com,a.com" {
		t.Fatalf("hosts = %v, %v", hosts, err)
	}
	for _, bad := range []*VMConfig{
		{Net: "allowlist", NetAllow: []string{"1.2.3.4"}},
		{Net: "open", NetAllow: []string{"a.com"}},
		{NetAllow: []string{"a.com"}},
	} {
		if _, err := bad.EffectiveNetAllow(); err == nil {
			t.Fatalf("accepted %+v", bad)
		}
	}
	if hosts, err := (*VMConfig)(nil).EffectiveNetAllow(); err != nil || hosts != nil {
		t.Fatalf("nil config = %v, %v", hosts, err)
	}
}

func TestAllowlistConfigYAMLRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "boite.yml")
	data, err := yaml.Marshal(BoiteConfig{VM: &VMConfig{Net: NetAllowlist, NetAllow: []string{"a.example.com"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadBoiteConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	hosts, err := cfg.VM.EffectiveNetAllow()
	if err != nil || len(hosts) != 1 || hosts[0] != "a.example.com" {
		t.Fatalf("round trip hosts = %v, %v", hosts, err)
	}
	if err := os.WriteFile(path, []byte("vm:\n  net: open\n  net_allow: [a.example.com]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadBoiteConfig(path); err == nil {
		t.Fatal("net_allow with open accepted")
	}
}

func TestAllowlistQEMUArgs(t *testing.T) {
	var cfg QEMUConfig
	applyVMConfig(&cfg, &BoiteConfig{VM: &VMConfig{Net: NetAllowlist, NetAllow: []string{"a.com"}}}, "/vm dir")
	if !cfg.RestrictNet || cfg.Passt != nil || cfg.Egress == nil {
		t.Fatalf("allowlist: RestrictNet=%v Passt=%v Egress=%v", cfg.RestrictNet, cfg.Passt, cfg.Egress)
	}
	cfg.Egress.Exe = "/opt/boite/boite"
	cfg.HostFwdPort = 2222
	args := BuildQEMUArgs(cfg)
	if !containsArg(args, "restrict=on") || !containsArg(args, "hostfwd=tcp:127.0.0.1:2222-:22") {
		t.Fatalf("args = %v", args)
	}
	want := "guestfwd=tcp:10.0.2.100:3128-cmd:'/opt/boite/boite' __egress-proxy '/vm dir/egress-hosts.txt' '/vm dir/egress.log'"
	if !containsArg(args, want) {
		t.Fatalf("missing %q in %v", want, args)
	}
	var other QEMUConfig
	applyVMConfig(&other, &BoiteConfig{VM: &VMConfig{Net: NetOffline}}, "/vm")
	if containsArg(BuildQEMUArgs(other), "guestfwd") {
		t.Fatal("offline must not carry a guestfwd")
	}
}

func TestEgressCommaEscape(t *testing.T) {
	e := &EgressConfig{Exe: "/a,b/boite", HostsFile: "/h", LogFile: "/l"}
	if !strings.Contains(e.guestfwd(), "'/a,,b/boite'") {
		t.Fatalf("comma not doubled: %s", e.guestfwd())
	}
}

func TestEgressPrepare(t *testing.T) {
	dir := t.TempDir()
	e := NewEgressConfig(dir, []string{"a.example.com"})
	if err := e.Prepare(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(e.HostsFile)
	if string(data) != "a.example.com\n" {
		t.Fatalf("hosts file = %q", data)
	}
	if err := (&EgressConfig{}).Prepare(); err == nil {
		t.Fatal("missing exe must fail closed")
	}
	if err := (*EgressConfig)(nil).Prepare(); err != nil {
		t.Fatal(err)
	}
}

func TestGuestProxyScript(t *testing.T) {
	on := GuestProxyScript(true)
	for _, s := range []string{"https_proxy=http://10.0.2.100:3128", "99boite-egress", "/etc/environment"} {
		if !strings.Contains(on, s) {
			t.Fatalf("enable script lacks %q: %s", s, on)
		}
	}
	off := GuestProxyScript(false)
	if strings.Contains(off, "10.0.2.100") || !strings.Contains(off, "rm -f") {
		t.Fatalf("disable script = %s", off)
	}
}
