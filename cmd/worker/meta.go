package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/FacileStudio/boite/cmd/qemu"
	"gopkg.in/yaml.v3"
)

type boxMeta struct {
	Owner  string `json:"owner"`
	CPUs   int    `json:"cpus"`
	Memory string `json:"memory"`
	Disk   string `json:"disk"`
	SSHKey string `json:"ssh_key,omitempty"`
	Net    string `json:"net,omitempty"`

	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type boxConfig struct {
	VM qemu.VMConfig `yaml:"vm"`
}

func metaPath(id string) string {
	return filepath.Join(qemu.GetInstanceDir(id), "worker.json")
}

func configPath(id string) string {
	return filepath.Join(qemu.GetInstanceDir(id), "boite.yml")
}

func saveMeta(id string, meta boxMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal box meta: %w", err)
	}
	return os.WriteFile(metaPath(id), data, 0o600)
}

func loadMeta(id string) (boxMeta, error) {
	var meta boxMeta
	data, err := os.ReadFile(metaPath(id))
	if err != nil {
		return meta, fmt.Errorf("read box meta: %w", err)
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return meta, fmt.Errorf("parse box meta: %w", err)
	}
	return meta, nil
}

func writeBoxFiles(req CreateRequest) error {
	if err := os.MkdirAll(qemu.GetInstanceDir(req.Name), 0o755); err != nil {
		return fmt.Errorf("create box dir: %w", err)
	}
	meta := boxMeta{
		Owner: req.Owner, CPUs: req.CPUs, Memory: req.Memory, Disk: req.Disk, SSHKey: req.SSHKey, Net: req.Net,
		Status: StatusCreating, CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeBoxConfig(req.Name, meta); err != nil {
		return err
	}
	return saveMeta(req.Name, meta)
}

func writeBoxConfig(id string, meta boxMeta) error {
	cfg := boxConfig{VM: qemu.VMConfig{CPUs: meta.CPUs, Memory: meta.Memory, Disk: meta.Disk, Net: meta.netMode()}}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal box config: %w", err)
	}
	if err := os.WriteFile(configPath(id), data, 0o644); err != nil {
		return fmt.Errorf("write box config: %w", err)
	}
	return nil
}

func (m boxMeta) netMode() string {
	if m.Net == "" {
		return NetOpen
	}
	return m.Net
}
