package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getSSHIdentityFile(inst *Instance) string {
	if inst.KeyPath != "" {
		return inst.KeyPath
	}

	if inst.PubKeyPath != "" {
		candidate := strings.TrimSuffix(inst.PubKeyPath, ".pub")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	home, _ := os.UserHomeDir()
	sshDir := filepath.Join(home, ".ssh")
	if _, err := os.Stat(filepath.Join(sshDir, "id_ed25519")); err == nil {
		return filepath.Join(sshDir, "id_ed25519")
	}
	if _, err := os.Stat(filepath.Join(sshDir, "id_rsa")); err == nil {
		return filepath.Join(sshDir, "id_rsa")
	}
	return filepath.Join(sshDir, "id_ed25519")
}

func BuildSSHArgs(inst *Instance, command []string) []string {
	args := []string{
		"-i", getSSHIdentityFile(inst),
		"-p", fmt.Sprintf("%d", inst.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ConnectTimeout=10",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
		"boite@127.0.0.1",
	}
	if len(command) > 0 {
		args = append(args, command...)
	}
	return args
}

func SSHCommand(inst *Instance, command []string) error {
	args := BuildSSHArgs(inst, command)
	sshCmd := exec.Command("ssh", args...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	err := sshCmd.Run()
	if err != nil {
		return fmt.Errorf("ssh -i %s -p %d boite@127.0.0.1: %w", getSSHIdentityFile(inst), inst.SSHPort, err)
	}
	return nil
}

func SSHInteractive(inst *Instance) error {
	return SSHCommand(inst, nil)
}
