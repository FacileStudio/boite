package qemu

import (
	"strings"
	"testing"
)

func TestValidateSnapshotTagAccepts(t *testing.T) {
	for _, tag := range []string{"clean", "run-1", "A_b-9", "a", strings.Repeat("x", 40)} {
		if err := validateSnapshotTag(tag); err != nil {
			t.Errorf("expected tag %q to be accepted, got %v", tag, err)
		}
	}
}

func TestValidateSnapshotTagRejects(t *testing.T) {
	for _, tag := range []string{"", strings.Repeat("x", 41), "has space", "has.dot", "slash/ed", "accenté", "tab\ttag"} {
		if err := validateSnapshotTag(tag); err == nil {
			t.Errorf("expected tag %q to be rejected", tag)
		}
	}
}
