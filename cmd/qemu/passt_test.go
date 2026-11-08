package qemu

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestPasstArgs(t *testing.T) {
	t.Setenv(SSHBindAddrEnv, "10.128.0.1")
	got := PasstArgs(NewPasstConfig("/vm"), 2223)
	want := []string{
		"--one-off", "--no-map-gw",
		"--socket", "/vm/passt.sock", "--pid", "/vm/passt.pid",
		"--tcp-ports", "10.128.0.1/2223:22",
		"--dns-forward", PasstDNS,
	}
	if !slices.Equal(got, want) {
		t.Fatalf("passt args:\n got %q\nwant %q", got, want)
	}
}

func TestBuildQEMUArgsPasstStream(t *testing.T) {
	args := BuildQEMUArgs(QEMUConfig{HostFwdPort: 2223, Passt: NewPasstConfig("/vm")})
	i := slices.Index(args, "-netdev")
	if i < 0 || args[i+1] != "stream,id=net0,server=off,addr.type=unix,addr.path=/vm/passt.sock" {
		t.Fatalf("expected passt stream netdev: %q", args)
	}
	if !slices.Contains(args, "virtio-net-pci,netdev=net0") || containsArg(args, "hostfwd=") {
		t.Fatalf("passt NIC must keep virtio-net-pci and drop slirp hostfwd: %q", args)
	}
}

func TestBuildQEMUArgsOfflineUnchanged(t *testing.T) {
	args := BuildQEMUArgs(QEMUConfig{HostFwdPort: 2223, RestrictNet: true})
	if !slices.Contains(args, "user,id=net0,restrict=on,hostfwd=tcp:127.0.0.1:2223-:22") {
		t.Fatalf("offline netdev changed: %q", args)
	}
}

func TestStartPasstMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := StartPasst(NewPasstConfig(t.TempDir()), 2223); err != ErrPasstMissing {
		t.Fatalf("expected ErrPasstMissing, got %v", err)
	}
	if err := StartPasst(nil, 2223); err != nil {
		t.Fatalf("nil passt config must be a no-op, got %v", err)
	}
}

func TestStopPasstSkipsForeignPID(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "passt.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := StopPasst(pidFile); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Fatal("pid file not removed")
	}
	if err := StopPasst(pidFile); err != nil {
		t.Fatalf("missing pid file must be a no-op, got %v", err)
	}
}

func TestStartPasstRejectsLongSocketPath(t *testing.T) {
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "passt"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	long := filepath.Join(t.TempDir(), strings.Repeat("d", 120))
	err := StartPasst(NewPasstConfig(long), 2223)
	if err == nil || !strings.Contains(err.Error(), "shorter BOITE_DIR") {
		t.Fatalf("expected socket path length error, got %v", err)
	}
}
