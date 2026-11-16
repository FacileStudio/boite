package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/FacileStudio/boite/cmd/qemu"
)

func TestOfflineBoxWritesRestrictedConfig(t *testing.T) {
	t.Setenv("BOITE_DIR", t.TempDir())
	b := NewBackend("10.0.0.1")
	req, err := CreateRequest{Name: "dev", Owner: "ann", Net: " Offline "}.Normalize()
	if err != nil {
		t.Fatal(err)
	}
	box, err := b.Reserve(req)
	if err != nil || box.Net != NetOffline {
		t.Fatalf("reserve: %+v %v", box, err)
	}
	if got := configNet(t, "dev"); got != NetOffline {
		t.Fatalf("boite.yml net = %q, want offline", got)
	}
	if _, err := b.SetNet("dev", NetOpen, nil); !errors.Is(err, ErrNotReady) {
		t.Fatalf("set net on creating = %v, want ErrNotReady", err)
	}
	if err := b.Fail(box, "boom"); err != nil {
		t.Fatal(err)
	}
	if got, _ := b.Get("dev"); got.Net != NetOffline {
		t.Fatalf("failed box lost its net: %+v", got)
	}
}

func TestWriteBoxConfigNet(t *testing.T) {
	t.Setenv("BOITE_DIR", t.TempDir())
	req, _ := CreateRequest{Name: "dev", Owner: "ann"}.Normalize()
	if err := writeBoxFiles(req); err != nil {
		t.Fatal(err)
	}
	if got := configNet(t, "dev"); got != NetOpen {
		t.Fatalf("default boite.yml net = %q, want open", got)
	}
	if err := writeBoxConfig("dev", boxMeta{CPUs: 2, Memory: "2G", Disk: "20G", Net: NetOffline}); err != nil {
		t.Fatal(err)
	}
	if got := configNet(t, "dev"); got != NetOffline {
		t.Fatalf("rewritten boite.yml net = %q, want offline", got)
	}
	if got := (boxMeta{}).netMode(); got != NetOpen {
		t.Fatalf("legacy meta net = %q, want open", got)
	}
}

func configNet(t *testing.T, id string) string {
	t.Helper()
	cfg, err := qemu.LoadBoiteConfig(configPath(id))
	if err != nil || cfg == nil || cfg.VM == nil {
		t.Fatalf("load config: %+v %v", cfg, err)
	}
	net, err := cfg.VM.EffectiveNet()
	if err != nil {
		t.Fatal(err)
	}
	return net
}

func TestNormalizeNetConfig(t *testing.T) {
	for _, net := range []string{"open-slirp", "lan"} {
		if _, _, err := NormalizeNetConfig(net, nil); !errors.Is(err, ErrInvalid) {
			t.Fatalf("net %q: expected ErrInvalid, got %v", net, err)
		}
	}
	if got, _, err := NormalizeNetConfig("", nil); err != nil || got != NetOpen {
		t.Fatalf("empty net = %q, %v", got, err)
	}
}

func TestNormalizeNetConfigAllowlist(t *testing.T) {
	mode, hosts, err := NormalizeNetConfig(" Allowlist ", []string{"API.Example.com", "*.example.org:8443", "api.example.com"})
	if err != nil || mode != NetAllowlist {
		t.Fatalf("allowlist = %q, %v", mode, err)
	}
	if len(hosts) != 2 || hosts[0] != "api.example.com" || hosts[1] != "*.example.org:8443" {
		t.Fatalf("hosts = %v", hosts)
	}
	if _, hosts, err := NormalizeNetConfig("allowlist", nil); err != nil || len(hosts) != 0 {
		t.Fatalf("allowlist without extras = %v, %v", hosts, err)
	}
	assertBadNetConfigs(t)
	tooMany := make([]string, 65)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf("h%d.example.com", i)
	}
	if _, _, err := NormalizeNetConfig("allowlist", tooMany); !errors.Is(err, ErrInvalid) {
		t.Fatalf("65 hosts: expected ErrInvalid, got %v", err)
	}
}

func TestAllowlistConfigRoundTrip(t *testing.T) {
	t.Setenv("BOITE_DIR", t.TempDir())
	req, err := CreateRequest{Name: "dev", Owner: "ann", Net: "allowlist", NetAllow: []string{"api.example.com"}}.Normalize()
	if err != nil {
		t.Fatal(err)
	}
	b := NewBackend("10.0.0.1")
	box, err := b.Reserve(req)
	if err != nil || box.Net != NetAllowlist || len(box.NetAllow) != 1 {
		t.Fatalf("reserve: %+v %v", box, err)
	}
	cfg, err := qemu.LoadBoiteConfig(configPath("dev"))
	if err != nil {
		t.Fatal(err)
	}
	hosts, err := cfg.VM.EffectiveNetAllow()
	if err != nil || len(hosts) != 1 || hosts[0] != "api.example.com" {
		t.Fatalf("boite.yml hosts = %v, %v", hosts, err)
	}
	meta, err := loadMeta("dev")
	if err != nil || meta.Net != NetAllowlist || meta.NetAllow[0] != "api.example.com" {
		t.Fatalf("worker.json = %+v, %v", meta, err)
	}
	if err := b.Fail(box, "boom"); err != nil {
		t.Fatal(err)
	}
	if got, _ := b.Get("dev"); got.Net != NetAllowlist || len(got.NetAllow) != 1 {
		t.Fatalf("failed box lost its allowlist: %+v", got)
	}
}

func TestLegacyMetaHasNoAllowlist(t *testing.T) {
	var meta boxMeta
	if err := json.Unmarshal([]byte(`{"owner":"a","cpus":2,"memory":"2G","disk":"20G"}`), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.netMode() != NetOpen || meta.NetAllow != nil || meta.GuestProxy {
		t.Fatalf("legacy meta = %+v", meta)
	}
}

func assertBadNetConfigs(t *testing.T) {
	t.Helper()
	bad := []struct {
		net   string
		allow []string
	}{
		{"allowlist", []string{"10.0.0.1"}},
		{"allowlist", []string{"localhost"}},
		{"allowlist", []string{"*.com"}},
		{"allowlist", []string{"a.com:0"}},
		{"open", []string{"a.com"}},
		{"offline", []string{"a.com"}},
	}
	for _, c := range bad {
		if _, _, err := NormalizeNetConfig(c.net, c.allow); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%q %v: expected ErrInvalid, got %v", c.net, c.allow, err)
		}
	}
}
