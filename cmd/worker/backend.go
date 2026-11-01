package worker

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/FacileStudio/boite/cmd/qemu"
)

// Backend is the production Lifecycle: it drives cmd/qemu directly and keeps
// each box's owner, resources and user key in worker.json beside state.json.
// A box that is being created or failed to create has worker.json only.
type Backend struct {
	host string
}

// NewBackend returns a Backend that reports host in every box's SSH info.
func NewBackend(host string) *Backend {
	return &Backend{host: host}
}

func (b *Backend) box(inst *qemu.Instance, meta boxMeta) Box {
	status := StatusStopped
	if inst.Status == StatusRunning && qemu.IsProcessRunning(inst.PID) {
		status = StatusRunning
	}
	return Box{
		ID: inst.Name, Name: inst.Name, Owner: meta.Owner, Status: status,
		CPUs: meta.CPUs, Memory: meta.Memory, Disk: meta.Disk,
		SSH:       SSHInfo{Host: b.host, Port: inst.SSHPort, User: sshUser},
		CreatedAt: inst.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (b *Backend) pendingBox(id string, meta boxMeta) Box {
	return Box{
		ID: id, Name: id, Owner: meta.Owner, Status: meta.Status,
		CPUs: meta.CPUs, Memory: meta.Memory, Disk: meta.Disk,
		SSH:       SSHInfo{Host: b.host, User: sshUser},
		CreatedAt: meta.CreatedAt, Error: meta.Error,
	}
}

func (b *Backend) readyBox(id string, meta boxMeta) (Box, error) {
	if meta.Status != "" {
		return b.pendingBox(id, meta), nil
	}
	inst, err := qemu.LoadInstanceState(id)
	if err != nil {
		return Box{}, err
	}
	return b.box(inst, meta), nil
}

func (b *Backend) load(id string) (*qemu.Instance, boxMeta, error) {
	meta, err := loadMeta(id)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, meta, ErrNotFound
	}
	if err != nil {
		return nil, meta, err
	}
	if meta.Status != "" {
		return nil, meta, ErrNotReady
	}
	inst, err := qemu.LoadInstanceState(id)
	if err != nil {
		return nil, meta, err
	}
	return inst, meta, nil
}

// Get returns one box with its live status.
func (b *Backend) Get(id string) (Box, error) {
	meta, err := loadMeta(id)
	if errors.Is(err, fs.ErrNotExist) {
		return Box{}, ErrNotFound
	}
	if err != nil {
		return Box{}, err
	}
	return b.readyBox(id, meta)
}

// List returns every box that has a worker.json, including boxes that are
// still being created or failed to create.
func (b *Backend) List() ([]Box, error) {
	entries, err := os.ReadDir(qemu.GetInstancesDir())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	boxes := make([]Box, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		meta, err := loadMeta(e.Name())
		if err != nil {
			continue
		}
		if box, err := b.readyBox(e.Name(), meta); err == nil {
			boxes = append(boxes, box)
		}
	}
	return boxes, nil
}

// Create provisions and boots a reserved box. A failed create leaves nothing
// behind; the caller records the failure with Fail.
func (b *Backend) Create(req CreateRequest) (Box, error) {
	var extra []string
	if req.SSHKey != "" {
		extra = []string{req.SSHKey}
	}
	inst, err := qemu.Create(qemu.CreateOptions{
		Name: req.Name, NoMount: true, GenerateKey: true, ConfigPath: configPath(req.Name),
		ExtraKeys: extra, Detached: true,
	})
	if err != nil {
		return Box{}, discard(req.Name, err)
	}
	meta, err := loadMeta(req.Name)
	if err != nil {
		return Box{}, discard(req.Name, err)
	}
	meta.Status, meta.Error = "", ""
	if err := saveMeta(req.Name, meta); err != nil {
		return Box{}, discard(req.Name, err)
	}
	return b.box(inst, meta), nil
}
