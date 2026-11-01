package worker

import (
	"errors"
	"io/fs"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
)

// Start boots a stopped box and re-applies its authorized keys, so a key or
// control-key change made while it was off takes effect.
func (b *Backend) Start(id string) (Box, error) {
	_, meta, err := b.load(id)
	if err != nil {
		return Box{}, err
	}
	inst, err := qemu.StartDetached(id, configPath(id))
	if err != nil {
		return Box{}, err
	}
	if err := qemu.ApplyAuthorizedKeys(inst, userKeys(meta)); err != nil {
		return Box{}, err
	}
	return b.box(inst, meta), nil
}

// Stop powers a box off and returns its stopped state.
func (b *Backend) Stop(id string) (Box, error) {
	if _, _, err := b.load(id); err != nil {
		return Box{}, err
	}
	if err := qemu.Stop(id); err != nil {
		return Box{}, err
	}
	return b.Get(id)
}

// Remove stops a box if needed and deletes all of its files.
func (b *Backend) Remove(id string) error {
	if _, err := os.Stat(qemu.GetInstanceDir(id)); errors.Is(err, fs.ErrNotExist) {
		return ErrNotFound
	}
	if qemu.InstanceExists(id) {
		return qemu.Destroy(id)
	}
	return qemu.DeleteInstanceDir(id)
}

func userKeys(meta boxMeta) []string {
	if meta.SSHKey == "" {
		return nil
	}
	return []string{meta.SSHKey}
}

// SetSSHKey replaces the box's user key. A running box gets it at once; a
// stopped box gets it on its next start.
func (b *Backend) SetSSHKey(id, key string) (Box, error) {
	inst, meta, err := b.load(id)
	if err != nil {
		return Box{}, err
	}
	if err := qemu.ValidatePublicKey(key); err != nil {
		return Box{}, invalid("ssh_key: %v", err)
	}
	meta.SSHKey = key
	if err := saveMeta(id, meta); err != nil {
		return Box{}, err
	}
	box := b.box(inst, meta)
	if box.Status == StatusRunning {
		if err := qemu.ApplyAuthorizedKeys(inst, userKeys(meta)); err != nil {
			return Box{}, err
		}
	}
	return box, nil
}
