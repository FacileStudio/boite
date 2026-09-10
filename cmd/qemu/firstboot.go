package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const firstbootMarker = "/var/lib/boite/firstboot.done"
const configVolID = "BOITECFG"

// BuildConfigISO writes the instance's SSH public key onto a tiny config ISO,
// the single per-instance input the baked base does not carry. The guest's
// boite-firstboot oneshot mounts this ISO on first boot, installs the key into
// /home/boite/.ssh/authorized_keys, then writes firstbootMarker. User
// creation, SSH host keys, package and toolchain install all live in the bake.
func BuildConfigISO(instanceDir, sshPubKey string) (string, error) {
	stageDir := filepath.Join(instanceDir, "config-stage")
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return "", fmt.Errorf("create config stage dir: %w", err)
	}

	keyPath := filepath.Join(stageDir, "authorized_keys")
	if err := os.WriteFile(keyPath, []byte(sshPubKey+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("write authorized_keys: %w", err)
	}

	configISOPath := filepath.Join(instanceDir, "config.iso")
	args := []string{"-quiet", "-r", "-volid", configVolID,
		"-o", configISOPath, keyPath}
	if _, err := exec.Command("genisoimage", args...).CombinedOutput(); err != nil {
		return "", fmt.Errorf("genisoimage: %w", err)
	}

	if _, err := os.Stat(configISOPath); err != nil {
		return "", fmt.Errorf("config.iso not created: %w", err)
	}
	if err := os.RemoveAll(stageDir); err != nil {
		return "", fmt.Errorf("clean config stage dir: %w", err)
	}
	return configISOPath, nil
}

// WaitForFirstboot blocks until the guest's boite-firstboot oneshot has
// published firstbootMarker. A connection or auth error means the guest is not
// ready (sshd comes up before the key lands), so every failed probe just keeps
// polling; only a marker-less timeout fails the create.
func WaitForFirstboot(inst *Instance, timeoutSeconds int) error {
	steps := timeoutSeconds / 5
	if steps < 1 {
		steps = 1
	}
	elapsed := 0
	for i := 0; i < steps; i++ {
		if firstbootDone(inst) {
			ProgressDone(fmt.Sprintf("firstboot complete (%ds)", elapsed))
			return nil
		}
		ProgressTick(fmt.Sprintf("firstboot (%ds): waiting for provisioned marker", elapsed), float64(i)/float64(steps))
		time.Sleep(5 * time.Second)
		elapsed += 5
	}
	ProgressFail("firstboot did not complete")
	return fmt.Errorf("firstboot: still not provisioned after %ds", timeoutSeconds)
}

// firstbootDone reports whether the guest's firstboot marker exists, treating
// an unreachable or still-authorizing guest as "not done yet".
func firstbootDone(inst *Instance) bool {
	cmd := exec.Command("ssh", BuildSSHArgs(inst, []string{"test", "-f", firstbootMarker})...)
	_, err := cmd.CombinedOutput()
	return err == nil
}
