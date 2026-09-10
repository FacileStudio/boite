package qemu

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

// Instance is the persisted state boite keeps for one sandbox: where its
// overlay and keys live, the SSH port, and whether the workspace is mounted.
type Instance struct {
	Name           string    `json:"name"`
	PID            int       `json:"pid"`
	SSHPort        int       `json:"ssh_port"`
	OverlayPath    string    `json:"overlay_path"`
	ConfigDiskPath string    `json:"config_disk_path"`
	KeyPath        string    `json:"key_path"`
	PubKeyPath     string    `json:"pub_key_path"`
	CreatedAt      time.Time `json:"created_at"`
	Status         string    `json:"status"`
	Workspace      string    `json:"workspace"`
	NoMount        bool      `json:"no_mount"`
	// PinnedEnv lists keys the user set via 'boite env set'. They are
	// authoritative: source refresh skips them, so the manual value sticks.
	PinnedEnv []string `json:"pinned_env,omitempty"`
}

// SaveInstanceState writes an instance's state as JSON to its state file,
// creating the instance directory first.
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

// LoadInstanceState reads an instance's state back from its state file.
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

// InstanceExists reports whether an instance has a state file on disk.
func InstanceExists(name string) bool {
	_, err := os.Stat(GetInstanceStatePath(name))
	return err == nil
}

// ListInstances returns every instance with state on disk, skipping any whose
// state fails to load. A missing instances directory is not an error.
func ListInstances() ([]*Instance, error) {
	entries, err := os.ReadDir(GetInstancesDir())
	if err != nil && !os.IsNotExist(err) {
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

// DeleteInstanceDir removes all of an instance's files, overlay included.
func DeleteInstanceDir(name string) error {
	return os.RemoveAll(GetInstanceDir(name))
}

// FindFreePort returns two consecutive free localhost ports, preferring lower
// numbers so restarting an instance keeps a stable SSH port over time.
func FindFreePort() (int, error) {
	for port := 2222; port <= 2322; port++ {
		if isPortFree(port) && isPortFree(port+1) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port in range 2222-2322")
}

// isPortFree reports whether a localhost TCP port is not accepting connections.
func isPortFree(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 50*time.Millisecond)
	if err == nil {
		conn.Close()
		return false
	}
	return true
}
