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

var syncOutDenyTestCases = []struct {
	name     string
	linkname string
	mode     int64
	typeflag byte
	want     string
}{
	{"main.go", "", 0o644, tar.TypeReg, ""},
	{"run.sh", "", 0o755, tar.TypeReg, ""},
	{".envrc", "", 0o644, tar.TypeReg, "matches deny list entry .envrc"},
	{"././.envrc", "", 0o644, tar.TypeReg, "matches deny list entry .envrc"},
	{".git/config", "", 0o644, tar.TypeReg, "matches deny list entry .git/config"},
	{".env.local", "", 0o644, tar.TypeReg, "matches deny list entry .env.*"},
	{".git/hooks/pre-commit", "", 0o755, tar.TypeReg, "matches deny list entry .git/hooks/*"},
	{"sub/../.git/hooks/pre-commit", "", 0o755, tar.TypeReg, "matches deny list entry .git/hooks/*"},
	{".git/hooks", "", 0o755, tar.TypeDir, ""},
	{".git/hooks", "/tmp/evil", 0o777, tar.TypeSymlink, "git directory cannot be a symlink"},
	{"tool", "", 0o4755, tar.TypeReg, "setuid bit set"},
	{"tool", "", 0o2755, tar.TypeReg, "setgid bit set"},
	{"scratch.txt", "", 0o666, tar.TypeReg, "world-writable"},
	{"link", "main.go", 0o777, tar.TypeSymlink, ""},
	{"link", "../../etc/passwd", 0o777, tar.TypeSymlink, "link target escapes workspace"},
	{"link", ".git/hooks/pre-commit", 0o777, tar.TypeSymlink, "link target matches deny list entry .git/hooks/*"},
	{"dir", "", 0o777, tar.TypeDir, ""},
	{"../../evil", "", 0o644, tar.TypeReg, "path traversal outside workspace"},
	{"devnode", "", 0o660, tar.TypeBlock, "special file not allowed"},
}

func TestSyncOutDenyReason(t *testing.T) {
	for _, tc := range syncOutDenyTestCases {
		hdr := &tar.Header{Name: tc.name, Linkname: tc.linkname, Mode: tc.mode, Typeflag: tc.typeflag}
		if got := syncOutDenyReason(hdr); got != tc.want {
			t.Errorf("syncOutDenyReason(%q, %q, %o, %q) = %q, want %q", tc.name, tc.linkname, tc.mode, string(tc.typeflag), got, tc.want)
		}
	}
}

func TestCleanSyncOutName(t *testing.T) {
	cases := []struct {
		name       string
		want       string
		wantEscape bool
	}{
		{"./.envrc", ".envrc", false},
		{".envrc", ".envrc", false},
		{"././.envrc", ".envrc", false},
		{".//.envrc", ".envrc", false},
		{"foo/../.envrc", ".envrc", false},
		{"./.git/hooks/x", ".git/hooks/x", false},
		{"./src/", "src", false},
		{"./", "", false},
		{"/abs/leet", "abs/leet", false},
		{"../../evil", "", true},
		{"/../../evil", "", true},
	}
	for _, tc := range cases {
		got, escape := cleanSyncOutName(tc.name)
		if got != tc.want || escape != tc.wantEscape {
			t.Errorf("cleanSyncOutName(%q) = (%q, %v), want (%q, %v)", tc.name, got, escape, tc.want, tc.wantEscape)
		}
	}
}
