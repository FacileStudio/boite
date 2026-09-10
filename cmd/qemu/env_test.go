package qemu

import (
	"reflect"
	"testing"
)

func TestEnvRefreshPlanSkipsPinned(t *testing.T) {
	got := envRefreshPlan(map[string]string{
		"FOO": "a",
		"BAR": "b",
		"BAZ": "c",
	}, []string{"BAR"})
	want := []string{"BAZ", "FOO"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("envRefreshPlan = %v, want %v", got, want)
	}
}

func TestEnvRefreshPlanAllPinned(t *testing.T) {
	got := envRefreshPlan(map[string]string{"FOO": "a"}, []string{"FOO"})
	if len(got) != 0 {
		t.Fatalf("expected no writes when the only key is pinned, got %v", got)
	}
}
