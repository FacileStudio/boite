package worker

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/FacileStudio/boite/cmd/qemu"
)

// Backend is the production Lifecycle: it drives cmd/qemu directly and keeps
// each box's owner, resources and user key in worker.json beside state.json.
type Backend struct {
	host     string
	createMu sync.Mutex
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

func (b *Backend) load(id string) (*qemu.Instance, boxMeta, error) {
	var meta boxMeta
	if !qemu.InstanceExists(id) {
		return nil, meta, ErrNotFound
	}
	inst, err := qemu.LoadInstanceState(id)
	if err != nil {
		return nil, meta, err
	}
	meta, err = loadMeta(id)
	if err != nil {
		return nil, meta, err
	}
	return inst, meta, nil
}

// Get returns one box with its live status.
func (b *Backend) Get(id string) (Box, error) {
	inst, meta, err := b.load(id)
	if err != nil {
		return Box{}, err
	}
	return b.box(inst, meta), nil
}

// List returns every box that has state on disk.
func (b *Backend) List() ([]Box, error) {
	instances, err := qemu.ListInstances()
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	boxes := make([]Box, 0, len(instances))
	for _, inst := range instances {
		if meta, err := loadMeta(inst.Name); err == nil {
			boxes = append(boxes, b.box(inst, meta))
		}
	}
	return boxes, nil
}

// Create provisions and boots a box. Creates run one at a time so two boxes
// never pick the same SSH port. A failed create leaves nothing behind.
func (b *Backend) Create(req CreateRequest) (Box, error) {
	b.createMu.Lock()
	defer b.createMu.Unlock()
	if qemu.InstanceExists(req.Name) {
		return Box{}, ErrExists
	}
	if err := writeBoxFiles(req); err != nil {
		return Box{}, discard(req.Name, err)
	}
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
		return Box{}, err
	}
	return b.box(inst, meta), nil
}

func discard(id string, cause error) error {
	if qemu.InstanceExists(id) {
		if err := qemu.Destroy(id); err != nil {
			slog.Warn("cleanup failed create: stop", "id", id, "err", err)
		}
	}
	if err := qemu.DeleteInstanceDir(id); err != nil {
		slog.Warn("cleanup failed create: delete", "id", id, "err", err)
	}
	return cause
}
