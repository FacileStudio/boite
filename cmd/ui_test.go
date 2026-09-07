package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/boite/cmd/qemu"
)

func TestStyleBanner(t *testing.T) {
	banner := "TEST BANNER"
	version := "0.1.6"
	out := styleBanner(banner, version)
	if !strings.Contains(out, banner) || !strings.Contains(out, version) {
		t.Fatalf("expected banner to contain banner text and version, got: %s", out)
	}
}

func TestRenderSandboxCard(t *testing.T) {
	name := "pingu"
	ws := "/home/yann/Code/test"
	sshPort := 22222
	card := renderSandboxCard(name, ws, false, sshPort)
	if !strings.Contains(card, name) || !strings.Contains(card, ws) {
		t.Fatalf("expected card to contain name and workspace, got: %s", card)
	}
	if !strings.Contains(card, "boite run pingu") {
		t.Fatalf("expected card to contain connect instruction, got: %s", card)
	}

	cardNoMount := renderSandboxCard(name, ws, true, sshPort)
	if strings.Contains(cardNoMount, ws) {
		t.Fatalf("expected card with noMount=true to not contain workspace, got: %s", cardNoMount)
	}
}

func TestRenderInstanceTableQEMU(t *testing.T) {
	instances := []*qemu.Instance{
		{Name: "pingu", Status: "running", SSHPort: 22222, CreatedAt: time.Now()},
		{Name: "test-box", Status: "stopped", SSHPort: 22223, CreatedAt: time.Now()},
	}
	rendered := renderInstanceTableQEMU(instances)
	if !strings.Contains(rendered, "pingu") || !strings.Contains(rendered, "test-box") {
		t.Fatalf("expected table to contain instances, got: %s", rendered)
	}
}