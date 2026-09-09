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

func SCPToInstance(inst *Instance, localPath, remotePath string) error {
	args := []string{
		"-i", getSSHIdentityFile(inst),
		"-P", fmt.Sprintf("%d", inst.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		localPath,
		fmt.Sprintf("boite@127.0.0.1:%s", remotePath),
	}
	scpCmd := exec.Command("scp", args...)
	scpCmd.Stdin = os.Stdin
	scpCmd.Stdout = os.Stdout
	scpCmd.Stderr = os.Stderr
	return scpCmd.Run()
}

func SCPFromInstance(inst *Instance, remotePath, localPath string) error {
	args := []string{
		"-i", getSSHIdentityFile(inst),
		"-P", fmt.Sprintf("%d", inst.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		fmt.Sprintf("boite@127.0.0.1:%s", remotePath),
		localPath,
	}
	scpCmd := exec.Command("scp", args...)
	scpCmd.Stdin = os.Stdin
	scpCmd.Stdout = os.Stdout
	scpCmd.Stderr = os.Stderr
	return scpCmd.Run()
}

func SyncWorkspaceZshrc(inst *Instance, localWorkspace string) error {
	zshrcPath := filepath.Join(localWorkspace, ".zshrc")
	if _, err := os.Stat(zshrcPath); os.IsNotExist(err) {
		return nil
	}
	return SCPToInstance(inst, zshrcPath, ".zshrc")
}
