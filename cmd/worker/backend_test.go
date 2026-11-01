package worker

import (
	"errors"
	"testing"
)

func TestBackendPendingRecords(t *testing.T) {
	t.Setenv("BOITE_DIR", t.TempDir())
	b := NewBackend("10.0.0.1")
	req, err := CreateRequest{Name: "dev", Owner: "ann"}.Normalize()
	if err != nil {
		t.Fatal(err)
	}
	box, err := b.Reserve(req)
	if err != nil || box.Status != StatusCreating || box.CreatedAt == "" || box.SSH.Host != "10.0.0.1" {
		t.Fatalf("reserve: %+v %v", box, err)
	}
	if _, err := b.Reserve(req); !errors.Is(err, ErrExists) {
		t.Fatalf("second reserve = %v, want ErrExists", err)
	}
	if _, err := b.Start("dev"); !errors.Is(err, ErrNotReady) {
		t.Fatalf("start on creating = %v, want ErrNotReady", err)
	}
	checkRecovered(t, b, box)
}

func checkRecovered(t *testing.T, b *Backend, box Box) {
	t.Helper()
	if err := b.Recover(); err != nil {
		t.Fatal(err)
	}
	got, err := b.Get("dev")
	if err != nil || got.Status != StatusError || got.Error != "interrupted" || got.CreatedAt != box.CreatedAt {
		t.Fatalf("after recover: %+v %v", got, err)
	}
	list, err := b.List()
	if err != nil || len(list) != 1 || list[0].ID != "dev" {
		t.Fatalf("list: %+v %v", list, err)
	}
	checkRemoved(t, b)
}

func checkRemoved(t *testing.T, b *Backend) {
	t.Helper()
	if err := b.Remove("dev"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Get("dev"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get after rm = %v", err)
	}
	if err := b.Remove("dev"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second rm = %v", err)
	}
}
