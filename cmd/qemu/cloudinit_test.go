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

func TestRenderUsesInstalledShellForBoiteUser(t *testing.T) {
	cfg := &CloudInitConfig{
		Users: []UserConfig{{
			Name:  "boite",
			Home:  "/home/boite",
			Shell: "/bin/zsh",
		}},
	}

	out := renderCloudInitUserData(cfg, "ssh-ed25519 test boite")

	if strings.Contains(out, "shell: /bin/zsh") {
		t.Fatal("boite user must not be created with an uninstalled shell (/bin/zsh): sshd rejects it and create fails")
	}
	if !strings.Contains(out, "shell: /bin/bash") {
		t.Fatal("boite user must be created with a base-image shell (/bin/bash) so provisioning SSH works")
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

func TestParseCloudInitError(t *testing.T) {
	raw := `{
  "errors": [
    "('scripts-user', RuntimeError('Runparts: 1 failures (runcmd) in 1 attempted commands'))"
  ],
  "status": "error"
}`
	module, message := parseCloudInitError(raw)
	if module != "scripts-user" {
		t.Errorf("parseCloudInitError module = %q, want scripts-user", module)
	}
	if !strings.Contains(message, "Runparts: 1 failures (runcmd)") {
		t.Errorf("parseCloudInitError message = %q, want runcmd failure text", message)
	}
}

func TestParseCloudInitErrorGarbage(t *testing.T) {
	module, message := parseCloudInitError("unrelated output")
	if module != "" || message != "" {
		t.Errorf("parseCloudInitError(%q) = %q, %q, want empty on non-error output", "unrelated output", module, message)
	}
}

func TestCloudInitFailingToken(t *testing.T) {
	if got := cloudInitFailingToken("zsh:1: command not found: facile"); got != "facile" {
		t.Errorf("cloudInitFailingToken command-not-found = %q, want facile", got)
	}
	if got := cloudInitFailingToken("sh: 1: nala: not found: abcd"); got != "abcd" {
		t.Errorf("cloudInitFailingToken bare not-found = %q, want abcd", got)
	}
	if got := cloudInitFailingToken("ordinary line"); got != "" {
		t.Errorf("cloudInitFailingToken on clean line = %q, want empty", got)
	}
}

func TestCloudInitConfigHintPointsAtCommandNotURL(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "boite.yml")
	contents := `# config
cloud_init:
  runcmd:
  - su - boite -c 'curl -fsSL https://get.facile.studio | bash'
  - su - boite -c 'facile install filet'
  - su - boite -c 'facile install sonde'
`
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	got := cloudInitConfigHint(p, "facile")
	if !strings.Contains(got, ":5:") {
		t.Errorf("cloudInitConfigHint = %q, want it to point at line 5 (the 'facile install filet' runcmd), not the URL line", got)
	}
	if !strings.Contains(got, "facile install filet") {
		t.Errorf("cloudInitConfigHint = %q, want the offending runcmd text", got)
	}
}

func TestIsFatalCloudInitState(t *testing.T) {
	if !isFatalCloudInitState("disabled") {
		t.Error("cloud-init 'disabled' must be fatal")
	}
	if !isFatalCloudInitState("failed") {
		t.Error("cloud-init 'failed' must be fatal")
	}
	if !isFatalCloudInitState("canceled") {
		t.Error("cloud-init 'canceled' must be fatal")
	}
	if isFatalCloudInitState("error") {
		t.Error("cloud-init 'error' must not be fatal: the boot completed, only a module failed")
	}
	if isFatalCloudInitState("running") {
		t.Error("cloud-init 'running' must not be fatal")
	}
	if isFatalCloudInitState("done") {
		t.Error("cloud-init 'done' must not be fatal")
	}
	if isFatalCloudInitState("") {
		t.Error("empty state must not be fatal")
	}
}
