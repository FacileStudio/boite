package qemu

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerateSSHKeyPairWritesInstanceScopedKeys(t *testing.T) {
	dir := t.TempDir()

	privPath, pubPath, err := GenerateSSHKeyPair(dir, "")
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

	if _, _, err := GenerateSSHKeyPair(dir1, ""); err != nil {
		t.Fatalf("first keygen failed: %v", err)
	}
	if _, _, err := GenerateSSHKeyPair(dir2, ""); err != nil {
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

	key1, _, _, _, err := ssh.ParseAuthorizedKey(pub1)
	if err != nil {
		t.Fatalf("parse first public key: %v", err)
	}
	key2, _, _, _, err := ssh.ParseAuthorizedKey(pub2)
	if err != nil {
		t.Fatalf("parse second public key: %v", err)
	}

	if key1.Type() != ssh.KeyAlgoED25519 || key2.Type() != ssh.KeyAlgoED25519 {
		t.Fatal("expected Ed25519 keys")
	}

	pubKey1, ok1 := key1.(ssh.PublicKey)
	pubKey2, ok2 := key2.(ssh.PublicKey)
	if !ok1 || !ok2 {
		t.Fatal("expected ssh.PublicKey")
	}
	if string(pubKey1.Marshal()) == string(pubKey2.Marshal()) {
		t.Fatal("expected different keys for different instance dirs")
	}
}

func TestGenerateSSHKeyPairWithPassphraseIsEncrypted(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := GenerateSSHKeyPair(dir, "hunter2"); err != nil {
		t.Fatalf("GenerateSSHKeyPair with passphrase failed: %v", err)
	}

	privBytes, err := os.ReadFile(filepath.Join(dir, "id_ed25519"))
	if err != nil {
		t.Fatalf("read private key: %v", err)
	}

	if _, err := ssh.ParseRawPrivateKey(privBytes); err == nil {
		t.Fatal("expected encrypted private key to fail parsing without passphrase")
	}

	if _, err := ssh.ParseRawPrivateKeyWithPassphrase(privBytes, []byte("hunter2")); err != nil {
		t.Fatalf("expected passphrase to decrypt key: %v", err)
	}
}

func TestReadPublicKey(t *testing.T) {
	dir := t.TempDir()
	_, pubPath, err := GenerateSSHKeyPair(dir, "")
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
