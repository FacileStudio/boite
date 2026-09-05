package cmd

import (
	"strings"
	"testing"
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
	card := renderSandboxCard(name, ws, false)
	if !strings.Contains(card, name) || !strings.Contains(card, ws) {
		t.Fatalf("expected card to contain name and workspace, got: %s", card)
	}
	if !strings.Contains(card, "boite shell pingu") {
		t.Fatalf("expected card to contain connect instruction, got: %s", card)
	}

	cardNoMount := renderSandboxCard(name, ws, true)
	if strings.Contains(cardNoMount, ws) {
		t.Fatalf("expected card with noMount=true to not contain workspace, got: %s", cardNoMount)
	}
}

func TestRenderInstanceTable(t *testing.T) {
	emptyJSON := []byte(`{"list": []}`)
	rendered, ok := renderInstanceTable(emptyJSON)
	if !ok {
		t.Fatal("expected ok=true for empty list")
	}
	if !strings.Contains(rendered, "No sandbox instances found") {
		t.Fatalf("expected empty message, got: %s", rendered)
	}

	validJSON := []byte(`{"list": [
		{"name": "pingu", "state": "Running", "ipv4": ["10.0.4.15"], "release": "Ubuntu 24.04 LTS"},
		{"name": "test-box", "state": "Stopped", "ipv4": [], "release": "Ubuntu 24.04 LTS"}
	]}`)
	rendered, ok = renderInstanceTable(validJSON)
	if !ok {
		t.Fatal("expected ok=true for valid list")
	}
	hasData := strings.Contains(rendered, "pingu") &&
		strings.Contains(rendered, "Running") &&
		strings.Contains(rendered, "10.0.4.15")
	if !hasData {
		t.Fatalf("expected table to contain instance data, got: %s", rendered)
	}

	invalidJSON := []byte(`not json at all`)
	_, ok = renderInstanceTable(invalidJSON)
	if ok {
		t.Fatal("expected ok=false for invalid json")
	}
}

func TestPrintHelpers(t *testing.T) {
	printSuccess("test success")
	printError("test error")
	printInfo("test info")
}
