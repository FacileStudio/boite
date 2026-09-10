package qemu

import (
	"os"
	"path/filepath"
)

const (
	InstancesDirName = "instances"
	CacheDirName     = "cache"
	// Base image is the baked boite image built by scripts/bake-image.sh. When
	// the image changes, rebuild and update the URL and the SHA256 together.
	BaseImageName   = "boite.qcow2"
	BaseImageURL    = "https://boite.facile.studio/base.qcow2"
	BaseImageSHA256 = "118a72ba67cd1a0f4f8e87a79faa9f06472a63e3eb8c58bf90797b23a1368f44"
)

// GetBoiteDir returns the root directory boite keeps all of its state under.
func GetBoiteDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".boite")
}

// GetInstancesDir returns the directory that holds state for every instance.
func GetInstancesDir() string {
	return filepath.Join(GetBoiteDir(), InstancesDirName)
}

// GetCacheDir returns the directory that holds the shared base image.
func GetCacheDir() string {
	return filepath.Join(GetBoiteDir(), CacheDirName)
}

// GetInstanceDir returns the directory that holds a single instance's files.
func GetInstanceDir(name string) string {
	return filepath.Join(GetInstancesDir(), name)
}

// GetBaseImagePath returns where the downloaded base image lives in the cache.
func GetBaseImagePath() string {
	return filepath.Join(GetCacheDir(), BaseImageName)
}