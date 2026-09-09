package qemu

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateSSHKeyPairWritesInstanceScopedKeys(t *testing.T) {
	dir := t.TempDir()

	privPath, pubPath, err := GenerateSSHKeyPair(dir)
	if err != nil {
		t.Fatalf("GenerateSSHKeyPair() failed: %v", err)
	}

	if privPath != filepath.Join(dir, "id_ed25519") {
		t.Errorf("expected private key at %s, got %s", filepath.Join(dir, "id_ed25519"), privPath)
	}
	if pubPath != filepath.Join(dir, "id_ed25519.pub") {
		t.Errorf("expected public key at %s, got %s", filepath.Join(dir, "id_ed25519.pub"), pubPath)
	}

	if _, err := os.Stat(privPath); err != nil {
		t.Fatalf("private key not written: %v", err)
	}
	if _, err := os.Stat(pubPath); err != nil {
		t.Fatalf("public key not written: %v", err)
	}

	if os.Getuid() == 0 {
		info, err := os.Stat(privPath)
		if err != nil {
			t.Fatalf("stat private key: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("expected private key mode 0600, got %04o", perm)
		}
	}
}

func TestGenerateSSHKeyPairUsesInstanceScopedPaths(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	if _, _, err := GenerateSSHKeyPair(dir1); err != nil {
		t.Fatalf("first keygen failed: %v", err)
	}
	if _, _, err := GenerateSSHKeyPair(dir2); err != nil {
		t.Fatalf("second keygen failed: %v", err)
	}

	pub1, err := os.ReadFile(filepath.Join(dir1, "id_ed25519.pub"))
	if err != nil {
		t.Fatalf("read first public key: %v", err)
	}
	pub2, err := os.ReadFile(filepath.Join(dir2, "id_ed25519.pub"))
	if err != nil {
		t.Fatalf("read second public key: %v", err)
	}

	if string(pub1) == string(pub2) {
		t.Fatal("expected different keys for different instance dirs")
	}
}

func TestReadPublicKey(t *testing.T) {
	dir := t.TempDir()
	_, pubPath, err := GenerateSSHKeyPair(dir)
	if err != nil {
		t.Fatalf("GenerateSSHKeyPair failed: %v", err)
	}

	key, err := ReadPublicKey(pubPath)
	if err != nil {
		t.Fatalf("ReadPublicKey failed: %v", err)
	}
	if !strings.HasPrefix(key, "ssh-ed25519 ") {
		t.Fatalf("expected ssh-ed25519 public key, got %q", key)
	}
}

func TestRenderDefersWriteFilesOwnedByBootedUser(t *testing.T) {
	cfg := &CloudInitConfig{
		Users: []UserConfig{{
			Name: "boite",
			Home: "/home/boite",
		}},
		WriteFiles: []WriteFileConfig{
			{Path: "/home/boite/.zshrc", Content: "alias ll='ls -la'\n", Owner: "boite:boite"},
			{Path: "/etc/motd", Content: "hi\n", Owner: "root:root"},
		},
	}

	out := renderCloudInitUserData(cfg, "ssh-ed25519 test boite")

	var got struct {
		WriteFiles []WriteFileConfig `yaml:"write_files"`
	}
	if err := yaml.Unmarshal([]byte(out[strings.Index(out, "\n")+1:]), &got); err != nil {
		t.Fatalf("unmarshal rendered user-data: %v", err)
	}

	deferred := map[string]bool{}
	for _, w := range got.WriteFiles {
		deferred[w.Path] = w.Defer
	}

	if !deferred["/home/boite/.zshrc"] {
		t.Fatal("expected boot-user-owned .zshrc to be deferred")
	}
	if deferred["/etc/motd"] {
		t.Fatal("root-owned /etc/motd must not be deferred")
	}
}

func TestCloudInitStateParsing(t *testing.T) {
	cases := []string{
		"status: running\n",
		"status: done",
		"status: error",
		"not reporting yet",
	}
	expect := []string{"running", "done", "error", ""}
	for i := 0; i < len(cases); i++ {
		if got := cloudInitState(cases[i]); got != expect[i] {
			t.Errorf("cloudInitState(%q) = %q, want %q", cases[i], got, expect[i])
		}
	}
}

func TestIsTerminalCloudInitError(t *testing.T) {
	if !isTerminalCloudInitError("error") {
		t.Error("cloud-init 'error' must be terminal")
	}
	if !isTerminalCloudInitError("disabled") {
		t.Error("cloud-init 'disabled' must be terminal")
	}
	if isTerminalCloudInitError("running") {
		t.Error("cloud-init 'running' must not be terminal")
	}
	if isTerminalCloudInitError("done") {
		t.Error("cloud-init 'done' must not be terminal")
	}
	if isTerminalCloudInitError("") {
		t.Error("empty state must not be terminal")
	}
}
