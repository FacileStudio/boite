package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SnapshotCreate records tag as an internal qcow2 snapshot on the instance's
// overlay disk. The VM must be stopped: running qemu holds a write lock on
// the overlay that qemu-img cannot take.
func SnapshotCreate(name, tag string) error {
	inst, err := stoppedInstance(name)
	if err != nil {
		return err
	}
	if err := validateSnapshotTag(tag); err != nil {
		return err
	}
	_, err = qemuImgSnapshot([]string{"snapshot", "-c", tag, inst.OverlayPath})
	return err
}

// SnapshotRollback applies tag to the instance's overlay disk, discarding
// everything written to the overlay since the snapshot was taken. The VM must
// be stopped.
func SnapshotRollback(name, tag string) error {
	inst, err := stoppedInstance(name)
	if err != nil {
		return err
	}
	if err := validateSnapshotTag(tag); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Warning: rolling back discards the current disk state of the sandbox")
	_, err = qemuImgSnapshot([]string{"snapshot", "-a", tag, inst.OverlayPath})
	return err
}

// SnapshotList renders the overlay's snapshot table with 'qemu-img snapshot
// -l'. The VM must be stopped.
func SnapshotList(name string) (string, error) {
	inst, err := stoppedInstance(name)
	if err != nil {
		return "", err
	}
	return qemuImgSnapshot([]string{"snapshot", "-l", inst.OverlayPath})
}

// stoppedInstance loads an instance and refuses it while its VM process is
// still alive.
func stoppedInstance(name string) (*Instance, error) {
	inst, err := LoadInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}
	if inst.PID > 0 && IsProcessRunning(inst.PID) {
		return nil, fmt.Errorf("sandbox '%s' is running: stop it first with 'boite stop %s'", name, name)
	}
	return inst, nil
}

// validateSnapshotTag rejects tags boite will not record: empty, longer than
// 40 characters, or containing anything outside letters, digits, dash and
// underscore.
func validateSnapshotTag(tag string) error {
	if len(tag) == 0 || len(tag) > 40 {
		return fmt.Errorf("invalid snapshot tag '%s': use 1 to 40 characters", tag)
	}
	for _, r := range tag {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return fmt.Errorf("invalid snapshot tag '%s': only letters, digits, dash and underscore are allowed", tag)
		}
	}
	return nil
}

// qemuImgSnapshot runs 'qemu-img <args>' and returns its combined output
// trimmed, failing with that output when qemu-img exits nonzero.
func qemuImgSnapshot(args []string) (string, error) {
	out, err := exec.Command("qemu-img", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("qemu-img %s: %s (%w)", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}
