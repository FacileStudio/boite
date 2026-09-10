package qemu

import "testing"

// shellQuote must keep a command's embedded spaces, pipes and quotes intact as
// a single sh -c argument once the remote shell re-parses the rejoined ssh
// command line. A pipeline is the classic case that a naked command breaks.
func TestShellQuoteKeepsPipelineIntact(t *testing.T) {
	cmd := `curl -fsSL "https://x.test/install.sh" | bash`
	want := `'curl -fsSL "https://x.test/install.sh" | bash'`
	if got := shellQuote(cmd); got != want {
		t.Fatalf("shellQuote(%q) = %q, want %q", cmd, got, want)
	}
}

func TestShellQuoteEscapesEmbeddedQuotes(t *testing.T) {
	cmd := `sh -lc 'export A="b"; echo $A'`
	want := `'sh -lc '\''export A="b"; echo $A'\'''`
	if got := shellQuote(cmd); got != want {
		t.Fatalf("shellQuote(%q) = %q, want %q", cmd, got, want)
	}
}
