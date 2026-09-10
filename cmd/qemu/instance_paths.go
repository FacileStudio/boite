package qemu

import (
	"path/filepath"
)

// GetInstanceStatePath returns the JSON state file for an instance.
func GetInstanceStatePath(name string) string {
	return filepath.Join(GetInstanceDir(name), "state.json")
}

// GetOverlayPath returns the copy-on-write overlay disk for an instance.
func GetOverlayPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "overlay.qcow2")
}

// GetConfigDiskPath returns the config vfat image for an instance.
func GetConfigDiskPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "config.img")
}

// GetKeyPath returns the private SSH key for an instance.
func GetKeyPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "id_ed25519")
}

// GetPubKeyPath returns the public SSH key for an instance.
func GetPubKeyPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "id_ed25519.pub")
}

// GetPIDPath returns the qemu PID file for an instance.
func GetPIDPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "qemu.pid")
}

// GetConsoleLogPath returns the console log for an instance.
func GetConsoleLogPath(name string) string {
	return filepath.Join(GetInstanceDir(name), "console.log")
}