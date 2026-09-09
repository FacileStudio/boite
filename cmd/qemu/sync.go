package qemu

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const workspaceDir = "/workspace"

// SyncWorkspaceIn replaces the sandbox's /workspace tree with a snapshot of
// srcDir, streamed as a tar archive over the instance's SSH transport. The
// remote side clears only /workspace first, so the host tree never leaks paths
// or permissions across the boundary: the guest only ever touches its own
// mounted-empty directory.
func SyncWorkspaceIn(inst *Instance, srcDir string) error {
	archive, err := captureCommand("tar", "-C", srcDir, "-czf", "-", ".")
	if err != nil {
		return fmt.Errorf("tar workspace: %w", err)
	}
	remote := fmt.Sprintf("mkdir -p %s && find %s -mindepth 1 -delete && tar -xzf - -C %s",
		workspaceDir, workspaceDir, workspaceDir)
	return sendArchive(inst, remote, archive)
}

// SyncWorkspaceOut copies /workspace from the sandbox back into destDir,
// overwriting matching files. Ownership from the guest (user boite) is not
// preserved so the copies stay owned by the host user.
func SyncWorkspaceOut(inst *Instance, destDir string) error {
	archive, err := captureRemote(inst, fmt.Sprintf("tar -czf - -C %s .", workspaceDir))
	if err != nil {
		return fmt.Errorf("tar workspace in VM: %w", err)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}
	untar := exec.Command("tar", "-xzf", "-", "-C", destDir, "--no-same-owner")
	untar.Stdin = bytes.NewReader(archive)
	if out, err := untar.CombinedOutput(); err != nil {
		return fmt.Errorf("untar to %s: %w: %s", destDir, err, out)
	}
	return nil
}

// captureCommand runs a local command and returns its stdout. Stderr is
// swallowed because tar is quiet unless it actually fails.
func captureCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// sendArchive pipes an in-memory archive into ssh stdin, where the remote
// command consumes it.
func sendArchive(inst *Instance, remote string, archive []byte) error {
	sshCmd := exec.Command("ssh", BuildSSHArgs(inst, []string{remote})...)
	sshCmd.Stdin = bytes.NewReader(archive)
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("workspace sync over ssh: %w", err)
	}
	return nil
}

// captureRemote runs a remote command over ssh and returns its stdout, which
// is the tar archive in the sync-out direction.
func captureRemote(inst *Instance, remote string) ([]byte, error) {
	sshCmd := exec.Command("ssh", BuildSSHArgs(inst, []string{remote})...)
	return sshCmd.Output()
}
