package worker

import (
	"errors"
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
	if _, err := b.SetNet("dev", NetOpen); !errors.Is(err, ErrNotReady) {
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
