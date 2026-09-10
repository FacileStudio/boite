package qemu

import (
	"bytes"
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

// BuildSSHArgs returns the ssh argv for the instance: the resolved identity
// file, the instance SSH port, and the flags that make the connection
// non-interactive and tolerant of a fresh guest host key. A non-empty command
// is appended as the remote command to run.
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

// WithEnv wraps a remote command so the guest's tiroir store is applied to its
// environment. A non-interactive ssh session never sources rc, so boite exec
// would silently lack vars without this. The store is eval'ed at the head of
// the command and the whole thing is returned as a single argument so it
// survives ssh's space-join/reparse intact. A missing tiroir binary becomes an
// empty eval, so the wrapped command still runs.
func WithEnv(command []string) []string {
	cmd := `eval "$(tiroir export)"; ` + strings.Join(command, " ")
	return []string{cmd}
}

// SSHCommand runs a one-shot command on the instance and propagates a nonzero
// remote exit status as an error.
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

// SSHOutput runs a remote command over ssh and returns its stdout. The command
// is a single shell string so it is not torn apart by ssh's space-join/reparse.
func SSHOutput(inst *Instance, command string) (string, error) {
	args := BuildSSHArgs(inst, []string{command})
	cmd := exec.Command("ssh", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ssh -i %s -p %d boite@127.0.0.1: %w", getSSHIdentityFile(inst), inst.SSHPort, err)
	}
	return out.String(), nil
}

// SSHInteractive opens a login shell. ssh reports a failed connection with
// exit 255; any other exit code (0 on a clean exit, or a nonzero remote shell
// status such as 127) means the session actually ran, so a dropped shell is
// not treated as an error.
func SSHInteractive(inst *Instance) error {
	sc := exec.Command("ssh", BuildSSHArgs(inst, nil)...)
	sc.Stdin = os.Stdin
	sc.Stdout = os.Stdout
	sc.Stderr = os.Stderr
	err := sc.Run()
	if err == nil {
		return nil
	}
	if sc.ProcessState == nil || sc.ProcessState.ExitCode() == 255 {
		return fmt.Errorf("ssh -i %s -p %d boite@127.0.0.1: %w", getSSHIdentityFile(inst), inst.SSHPort, err)
	}
	return nil
}
