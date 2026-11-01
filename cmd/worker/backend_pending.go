package worker

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
)

// Reserve persists a box in status creating and returns it. It fails with
// ErrExists when the name is taken.
func (b *Backend) Reserve(req CreateRequest) (Box, error) {
	if _, err := os.Stat(qemu.GetInstanceDir(req.Name)); err == nil {
		return Box{}, ErrExists
	}
	if err := writeBoxFiles(req); err != nil {
		return Box{}, discard(req.Name, err)
	}
	meta, err := loadMeta(req.Name)
	if err != nil {
		return Box{}, discard(req.Name, err)
	}
	return b.pendingBox(req.Name, meta), nil
}

// Fail replaces whatever is left of a box with an error record carrying msg.
func (b *Backend) Fail(box Box, msg string) error {
	discard(box.ID, nil)
	if err := os.MkdirAll(qemu.GetInstanceDir(box.ID), 0o755); err != nil {
		return fmt.Errorf("create box dir: %w", err)
	}
	return saveMeta(box.ID, boxMeta{
		Owner: box.Owner, CPUs: box.CPUs, Memory: box.Memory, Disk: box.Disk,
		Status: StatusError, Error: msg, CreatedAt: box.CreatedAt,
	})
}

// Recover marks every box left in status creating by a previous worker as
// failed with the message "interrupted".
func (b *Backend) Recover() error {
	entries, err := os.ReadDir(qemu.GetInstancesDir())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("list instances: %w", err)
	}
	for _, e := range entries {
		meta, err := loadMeta(e.Name())
		if err != nil || meta.Status != StatusCreating {
			continue
		}
		if err := b.Fail(b.pendingBox(e.Name(), meta), "interrupted"); err != nil {
			return err
		}
	}
	return nil
}

func discard(id string, cause error) error {
	if qemu.InstanceExists(id) {
		if err := qemu.Destroy(id); err != nil {
			slog.Warn("clean up failed create: stop", "id", id, "err", err)
		}
	}
	if err := qemu.DeleteInstanceDir(id); err != nil {
		slog.Warn("clean up failed create: delete", "id", id, "err", err)
	}
	return cause
}
