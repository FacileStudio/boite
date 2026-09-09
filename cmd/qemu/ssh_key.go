package qemu

import (
	"fmt"
)

type sshKeyResolution struct {
	pubKey         string
	privateKeyPath string
	publicKeyPath  string
}

func resolveSSHKey(instanceDir string, generateKey bool, cfg *BoiteConfig) (sshKeyResolution, error) {
	var resolution sshKeyResolution
	var err error

	if generateKey {
		resolution.privateKeyPath, resolution.publicKeyPath, err = GenerateSSHKeyPair(instanceDir)
		if err != nil {
			return resolution, fmt.Errorf("generate ssh keys: %w", err)
		}
		resolution.pubKey, err = ReadPublicKey(resolution.publicKeyPath)
		if err != nil {
			return resolution, fmt.Errorf("read public key: %w", err)
		}
		return resolution, nil
	}

	resolution.pubKey, err = FindUserSSHKey()
	if err != nil {
		return resolution, fmt.Errorf("find user ssh key: %w\nHint: run 'ssh-keygen -t ed25519' to create one, or use --generate-key to create a per-VM key", err)
	}
	resolution.privateKeyPath, err = FindUserSSHPrivateKey()
	if err != nil {
		return resolution, fmt.Errorf("find user ssh private key: %w\nHint: run 'ssh-keygen -t ed25519' to create one, or use --generate-key to create a per-VM key", err)
	}
	return resolution, nil
}
