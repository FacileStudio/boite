package qemu

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

// GenerateSeedISO creates the seed.iso from the merged cloud-init config.
// configPath is passed through so explicit `--config` files are honored.
func GenerateSeedISO(instanceDir, instanceName, sshPubKey, configPath string) (string, error) {
	cfg, err := LoadBoiteConfig(configPath)
	if err != nil {
		return "", fmt.Errorf("load boite config: %w", err)
	}

	mergedCfg := MergeCloudInitConfig(cfg)
	if mergedCfg == nil {
		mergedCfg = &CloudInitConfig{}
	}

	userExists := false
	for _, user := range mergedCfg.Users {
		if user.Name == "boite" {
			userExists = true
			if !slices.Contains(user.SSHAuthorizedKeys, sshPubKey) {
				user.SSHAuthorizedKeys = append(user.SSHAuthorizedKeys, sshPubKey)
			}
			break
		}
	}

	if !userExists {
		mergedCfg.Users = append(mergedCfg.Users, UserConfig{
			Name:              "boite",
			Gecos:             "Boite",
			Groups:            []string{"sudo"},
			Home:              "/home/boite",
			Shell:             "/bin/zsh",
			Sudo:              "ALL=(ALL) NOPASSWD:ALL",
			SSHAuthorizedKeys: []string{sshPubKey},
			Password:          "boite",
		})
	}

	userData := renderCloudInitUserData(mergedCfg, sshPubKey)
	metaData := fmt.Sprintf("instance-id: boite-%s\nlocal-hostname: boite\n", instanceName)

	userDataPath := filepath.Join(instanceDir, "user-data")
	metaDataPath := filepath.Join(instanceDir, "meta-data")
	seedISOPath := filepath.Join(instanceDir, "seed.iso")

	if err := writeCloudInitFiles(userDataPath, userData, metaDataPath, metaData); err != nil {
		return "", err
	}

	if err := runCloudLocalDS(seedISOPath, userDataPath, metaDataPath); err != nil {
		return "", err
	}

	if _, err := os.Stat(seedISOPath); err != nil {
		return "", fmt.Errorf("seed.iso not created: %w", err)
	}

	return seedISOPath, nil
}

// anyPasswordInUsers returns true if any user in the cloud-init config has a password set
func anyPasswordInUsers(cfg map[string]any) bool {
	if cfg == nil || cfg["users"] == nil {
		return false
	}
	users := cfg["users"].([]map[string]any)
	for _, umap := range users {
		if _, has := umap["password"]; has {
			return true
		}
	}
	return false
}

// writeCloudInitFiles writes the user-data and meta-data files for cloud-init
func writeCloudInitFiles(userDataPath, userData, metaDataPath, metaData string) error {
	if err := os.WriteFile(userDataPath, []byte(userData), 0o644); err != nil {
		return fmt.Errorf("write user-data: %w", err)
	}
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0o644); err != nil {
		return fmt.Errorf("write meta-data: %w", err)
	}
	return nil
}

// runCloudLocalDS creates a seed ISO using cloud-localds
func runCloudLocalDS(seedISOPath, userDataPath, metaDataPath string) error {
	args := []string{"cloud-localds", seedISOPath, userDataPath, metaDataPath}
	cmd := exec.Command(args[0], args[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cloud-localds: %w", err)
	}
	return nil
}

// ComputeInstanceID generates a short instance ID from the name
func ComputeInstanceID(name string) string {
	h := sha256.Sum256([]byte(name + "boite-instance"))
	return fmt.Sprintf("%x", h[:8])
}
