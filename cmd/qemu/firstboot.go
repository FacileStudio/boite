package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tiroir "github.com/FacileStudio/tiroir/lib"
)

const firstbootMarker = "/var/lib/boite/firstboot.done"
const configVolID = "BOITECFG"

// BuildConfigDisk writes the instance's SSH public key onto a tiny vfat disk
// image, the single per-instance input the baked base does not carry. The
// guest's boite-firstboot oneshot mounts this disk on first boot, installs the
// key into /home/boite/.ssh/authorized_keys, then writes firstbootMarker. User
// creation, SSH host keys, package and toolchain install all live in the bake.
//
// When envVars is non-empty, the helper also initializes an encrypted tiroir
// store with those values and copies the ciphertext blob plus its key onto the
// disk, so the guest's store exists before its first command. Only the encrypted
// .tiroir and .tiroir.key are written — never plaintext, never a host token.
func BuildConfigDisk(instanceDir, sshPubKey string, envVars map[string]string) (string, error) {
	stageDir := filepath.Join(instanceDir, "config-stage")
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return "", fmt.Errorf("create config stage dir: %w", err)
	}

	diskPath := filepath.Join(instanceDir, "config.img")
	if _, err := exec.Command("truncate", "-s", "4M", diskPath).CombinedOutput(); err != nil {
		return "", fmt.Errorf("size config disk: %w", err)
	}

	if _, err := exec.Command(mkfsFat(), mkfsFatOpts(diskPath)...).CombinedOutput(); err != nil {
		return "", fmt.Errorf("mkfs.fat: %w", err)
	}

	keyPath := filepath.Join(stageDir, "authorized_keys")
	if err := os.WriteFile(keyPath, []byte(sshPubKey+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("write authorized_keys: %w", err)
	}
	if _, err := exec.Command("mcopy", []string{"-i", diskPath, keyPath, "::/authorized_keys"}...).CombinedOutput(); err != nil {
		return "", fmt.Errorf("mcopy: %w", err)
	}

	if err := writeTiroirPayload(diskPath, stageDir, envVars); err != nil {
		return "", err
	}

	if err := os.RemoveAll(stageDir); err != nil {
		return "", fmt.Errorf("clean config stage dir: %w", err)
	}
	return diskPath, nil
}

// writeTiroirPayload initializes an encrypted tiroir store in stageDir with
// envVars and copies the ciphertext blob plus its key onto the config disk.
// An empty map writes no payload. The split keyfile means a store without its
// key cannot decrypt, so the guest never sees plaintext env on disk or wire.
func writeTiroirPayload(diskPath, stageDir string, envVars map[string]string) error {
	if len(envVars) == 0 {
		return nil
	}
	store := tiroir.NewAt(stageDir)
	for k, v := range envVars {
		if err := store.Set(k, v); err != nil {
			return fmt.Errorf("tiroir set %s: %w", k, err)
		}
	}
	for _, name := range []string{".tiroir", ".tiroir.key"} {
		src := filepath.Join(stageDir, name)
		if _, err := exec.Command("mcopy", []string{"-i", diskPath, src, "::/" + name}...).CombinedOutput(); err != nil {
			return fmt.Errorf("mcopy %s: %w", name, err)
		}
	}
	return nil
}

// mkfsFat resolves the mkfs.fat binary across PATH and common /sbin locations.
func mkfsFat() string {
	if _, err := exec.LookPath("mkfs.fat"); err == nil {
		return "mkfs.fat"
	}
	return "/sbin/mkfs.fat"
}

// mkfsFatOpts builds the mkfs.fat invocation for the config volume.
func mkfsFatOpts(diskPath string) []string {
	return []string{"-n", configVolID, diskPath}
}

// WaitForFirstboot blocks until the guest's boite-firstboot oneshot has
// published firstbootMarker. A connection or auth error means the guest is not
// ready (sshd comes up before the key lands), so every failed probe just keeps
// polling; only a marker-less timeout fails the create.
func WaitForFirstboot(inst *Instance, timeoutSeconds int) error {
	steps := max(timeoutSeconds/5, 1)
	elapsed := 0
	for i := range steps {
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
