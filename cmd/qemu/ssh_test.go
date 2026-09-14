package qemu

import "testing"

func TestWithEnv(t *testing.T) {
	got := WithEnv([]string{"echo", "hello world"})
	if len(got) != 1 {
		t.Fatalf("expected a single ssh argument, got %d", len(got))
	}
	want := `export PATH="/usr/local/go/bin:$HOME/.cargo/bin:$HOME/.bun/bin:$HOME/.local/bin:/usr/local/bin:$PATH"; eval "$(tiroir export 2>/dev/null || true)"; 'echo' 'hello world'`
	if got[0] != want {
		t.Fatalf("WithEnv = %q, want %q", got[0], want)
	}
}
