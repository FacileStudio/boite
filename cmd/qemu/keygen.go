package qemu

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	keygen "github.com/charmbracelet/keygen"
)

// GenerateSSHKeyPair creates an Ed25519 key pair using keygen and writes
// them to the instance directory. Existing keys are reused if found on disk.
func GenerateSSHKeyPair(instanceDir string) (string, string, error) {
	keyPath := filepath.Join(instanceDir, "id_ed25519")
	opts := []keygen.Option{
		keygen.WithKeyType(keygen.Ed25519),
		keygen.WithWrite(),
	}

	kp, err := keygen.New(keyPath, opts...)
	if err != nil {
		return "", "", fmt.Errorf("keygen: %w", err)
	}

	privatePath := keyPath
	publicPath := keyPath + ".pub"
	privBytes := kp.RawPrivateKey()
	if err := os.WriteFile(privatePath, privBytes, 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
	}
	if err := os.Chmod(privatePath, 0o600); err != nil {
		return "", "", fmt.Errorf("chmod private key: %w", err)
	}

	pubBytes := []byte(kp.AuthorizedKey())
	if err := os.WriteFile(publicPath, pubBytes, 0o644); err != nil {
		return "", "", fmt.Errorf("write public key: %w", err)
	}

	return privatePath, publicPath, nil
}

// ReadPublicKey reads the public key file and returns the trimmed key string
func ReadPublicKey(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
