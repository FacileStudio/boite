package qemu

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	InstancesDirName = "instances"
	CacheDirName     = "cache"
	BaseImageName    = "debian-12-genericcloud-amd64.qcow2"
	BaseImageURL     = "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2"
)

type Instance struct {
	Name        string    `json:"name"`
	PID         int       `json:"pid"`
	SSHPort     int       `json:"ssh_port"`
	OverlayPath string    `json:"overlay_path"`
	SeedISOPath string    `json:"seed_iso_path"`
	KeyPath     string    `json:"key_path"`
	PubKeyPath  string    `json:"pub_key_path"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	Workspace   string    `json:"workspace"`
	NoMount     bool      `json:"no_mount"`
}

func GetBoiteDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".boite")
}

func GetInstancesDir() string {
	return filepath.Join(GetBoiteDir(), InstancesDirName)
}

func GetCacheDir() string {
	return filepath.Join(GetBoiteDir(), CacheDirName)
}

func GetInstanceDir(name string) string {
	return filepath.Join(GetInstancesDir(), name)
}

func GetInstanceStatePath(name string) string {
	return filepath.Join(GetInstanceDir(name), "state.json")
}

func GetOverlayPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "overlay.qcow2")
}

func GetSeedISOPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "seed.iso")
}

func GetKeyPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "id_ed25519")
}

func GetPubKeyPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "id_ed25519.pub")
}

func GetPIDPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "qemu.pid")
}

func GetBaseImagePath() string {
	return filepath.Join(GetCacheDir(), BaseImageName)
}

func SaveInstanceState(inst *Instance) error {
	dir := GetInstanceDir(inst.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create instance dir: %w", err)
	}
	data, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return os.WriteFile(GetInstanceStatePath(inst.Name), data, 0o644)
}

func LoadInstanceState(name string) (*Instance, error) {
	data, err := os.ReadFile(GetInstanceStatePath(name))
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	var inst Instance
	if err := json.Unmarshal(data, &inst); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}
	return &inst, nil
}

func InstanceExists(name string) bool {
	_, err := os.Stat(GetInstanceStatePath(name))
	return err == nil
}

func ListInstances() ([]*Instance, error) {
	entries, err := os.ReadDir(GetInstancesDir())
	if err != nil {
		return nil, err
	}
	var instances []*Instance
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		inst, err := LoadInstanceState(e.Name())
		if err != nil {
			continue
		}
		instances = append(instances, inst)
	}
	return instances, nil
}

func DeleteInstanceDir(name string) error {
	return os.RemoveAll(GetInstanceDir(name))
}

func FindFreePort() (int, error) {
	for port := 2222; port <= 2322; port++ {
		if isPortFree(port) && isPortFree(port+1) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port in range 2222-2322")
}

func isPortFree(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 50*time.Millisecond)
	if err == nil {
		conn.Close()
		return false
	}
	return true
}