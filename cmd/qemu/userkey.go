package qemu

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindUserSSHKey returns the contents of the user's first readable public SSH
// key (id_ed25519.pub, then id_rsa.pub, then id_dsa.pub), trimmed of
// surrounding whitespace.
func FindUserSSHKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	keyFiles := []string{
		filepath.Join(home, ".ssh", "id_ed25519.pub"),
		filepath.Join(home, ".ssh", "id_rsa.pub"),
		filepath.Join(home, ".ssh", "id_dsa.pub"),
	}

	var lastError error
	for _, keyFile := range keyFiles {
		if _, err := os.Stat(keyFile); err != nil {
			lastError = err
			continue
		}
		data, err := os.ReadFile(keyFile)
		if err != nil {
			lastError = err
			continue
		}
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("no user SSH key found: %w", lastError)
}

// FindUserSSHPrivateKey returns the path of the user's first existing private
// SSH key (id_ed25519, then id_rsa, then id_dsa).
func FindUserSSHPrivateKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	keyFiles := []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_dsa"),
	}

	var lastError error
	for _, keyFile := range keyFiles {
		if _, err := os.Stat(keyFile); err != nil {
			lastError = err
			continue
		}
		return keyFile, nil
	}

	return "", fmt.Errorf("no user ssh private key found: %w", lastError)
}
