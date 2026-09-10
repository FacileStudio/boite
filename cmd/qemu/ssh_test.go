package qemu

import "testing"

func TestWithEnv(t *testing.T) {
	got := WithEnv([]string{"echo", "hi"})
	if len(got) != 1 {
		t.Fatalf("expected a single ssh argument, got %d", len(got))
	}
	want := `eval "$(tiroir export)"; echo hi`
	if got[0] != want {
		t.Fatalf("WithEnv = %q, want %q", got[0], want)
	}
}
