package qemu

import (
	"archive/tar"
	"testing"
)

func TestMatchSyncOutDeny(t *testing.T) {
	cases := []struct {
		pattern string
		name    string
		want    bool
	}{
		{".git/hooks/*", ".git/hooks/pre-commit", true},
		{".git/hooks/*", ".git/hooks/pre-commit.d/sample", true},
		{".git/hooks/*", ".git/hooks", false},
		{".git/hooks/*", ".gitignore", false},
		{".git/hooks/*", "src/.git/hooks/pre-commit", false},
		{".envrc", ".envrc", true},
		{".envrc", "sub/.envrc", false},
		{".env", ".env", true},
		{".env", ".envrc", false},
		{".env.*", ".env.local", true},
		{".env.*", ".env.production.secret", true},
		{".env.*", ".env", false},
		{".env.*", "deploy/.env.staging", false},
		{".vscode/tasks.json", ".vscode/tasks.json", true},
		{".vscode/tasks.json", ".vscode/settings.json", false},
		{".husky/*", ".husky/pre-commit", true},
		{".husky/*", ".husky/_/husky.sh", true},
		{".husky/*", ".husky", false},
	}
	for _, tc := range cases {
		if got := matchSyncOutDeny(tc.pattern, tc.name); got != tc.want {
			t.Errorf("matchSyncOutDeny(%q, %q) = %v, want %v", tc.pattern, tc.name, got, tc.want)
		}
	}
}

func TestSyncOutDenyReason(t *testing.T) {
	cases := []struct {
		name     string
		mode     int64
		typeflag byte
		want     string
	}{
		{"main.go", 0o644, tar.TypeReg, ""},
		{"run.sh", 0o755, tar.TypeReg, ""},
		{".envrc", 0o644, tar.TypeReg, "matches deny list entry .envrc"},
		{".env.local", 0o644, tar.TypeReg, "matches deny list entry .env.*"},
		{".git/hooks/pre-commit", 0o755, tar.TypeReg, "matches deny list entry .git/hooks/*"},
		{".git/hooks", 0o755, tar.TypeDir, ""},
		{"tool", 0o4755, tar.TypeReg, "setuid bit set"},
		{"tool", 0o2755, tar.TypeReg, "setgid bit set"},
		{"scratch.txt", 0o666, tar.TypeReg, "world-writable"},
		{"link", 0o777, tar.TypeSymlink, ""},
		{"dir", 0o777, tar.TypeDir, ""},
	}
	for _, tc := range cases {
		hdr := &tar.Header{Name: tc.name, Mode: tc.mode, Typeflag: tc.typeflag}
		if got := syncOutDenyReason(hdr); got != tc.want {
			t.Errorf("syncOutDenyReason(%q, %o, %q) = %q, want %q", tc.name, tc.mode, string(tc.typeflag), got, tc.want)
		}
	}
}

func TestCleanSyncOutName(t *testing.T) {
	cases := map[string]string{
		"./.envrc":       ".envrc",
		".envrc":         ".envrc",
		"./.git/hooks/x": ".git/hooks/x",
		"./src/":         "src",
		"./":             "",
		"/abs/leet":      "abs/leet",
	}
	for name, want := range cases {
		if got := cleanSyncOutName(name); got != want {
			t.Errorf("cleanSyncOutName(%q) = %q, want %q", name, got, want)
		}
	}
}
