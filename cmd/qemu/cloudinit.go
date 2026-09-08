package qemu

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	keygen "github.com/charmbracelet/keygen"
	"gopkg.in/yaml.v3"
)

// GenerateSSHKeyPair creates an Ed25519 key pair using keygen and writes
// them to the central key store at ~/.boite/ssh/. If a passphrase is
// provided, the private key is encrypted at rest. Existing keys are
// reused if found on disk.
func GenerateSSHKeyPair(instanceDir string, passphrase string) (string, string, error) {
	keyPath := filepath.Join(instanceDir, "id_ed25519")
	opts := []keygen.Option{
		keygen.WithKeyType(keygen.Ed25519),
		keygen.WithWrite(),
	}
	if passphrase != "" {
		opts = append(opts, keygen.WithPassphrase(passphrase))
	}

	kp, err := keygen.New(keyPath, opts...)
	if err != nil {
		return "", "", fmt.Errorf("keygen: %w", err)
	}

	// Calculate paths based on the keyPath we passed in
	privatePath := keyPath
	publicPath := keyPath + ".pub"

	// Write the private key (potentially encrypted) to disk
	privBytes := kp.RawProtectedPrivateKey()
	if privBytes == nil {
		privBytes = kp.RawPrivateKey()
	}
	if err := os.WriteFile(privatePath, privBytes, 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
	}
	if err := os.Chmod(privatePath, 0o600); err != nil {
		return "", "", fmt.Errorf("chmod private key: %w", err)
	}

	// Write public key
	pubBytes := []byte(kp.AuthorizedKey())
	if err := os.WriteFile(publicPath, pubBytes, 0o644); err != nil {
		return "", "", fmt.Errorf("write public key: %w", err)
	}

	return privatePath, publicPath, nil
}

// ReadPublicKey reads the public key file and returns the trimmed key string
func ReadPublicKey(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

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

// renderCloudInitUserData builds the cloud-init user-data YAML from the config
func renderCloudInitUserData(cfg *CloudInitConfig, sshPubKey string) string {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	cloudConfig := buildBaseCloudConfig()
	addUsersToConfig(cloudConfig, cfg.Users, sshPubKey)
	addPackagesToConfig(cloudConfig, cfg.Packages)
	addAPTSourcesToConfig(cloudConfig, cfg.APT)
	addRuncmdToConfig(cloudConfig, cfg.Runcmd)
	addWriteFilesToConfig(cloudConfig, cfg.WriteFiles)

	if err := encoder.Encode(cloudConfig); err != nil {
		return "#cloud-config\n" + buf.String()
	}
	if err := encoder.Close(); err != nil {
		return "#cloud-config\n" + buf.String()
	}

	return "#cloud-config\n" + buf.String()
}

func buildBaseCloudConfig() map[string]any {
	eth0 := map[string]any{
		"dhcp4":       false,
		"addresses":   []string{"192.168.42.10/24"},
		"routes":      []map[string]any{{"to": "default", "via": "192.168.42.1"}},
		"nameservers": map[string]any{"addresses": []string{"8.8.8.8", "8.8.4.4"}},
	}
	ethernets := map[string]any{"eth0": eth0}
	network := map[string]any{"version": 2, "ethernets": ethernets}

	return map[string]any{
		"hostname":       "boite",
		"disable_root":   true,
		"ssh_pwauth":     false,
		"package_update": true,
		"network":        network,
	}
}

func addUsersToConfig(cloudConfig map[string]any, users []UserConfig, sshPubKey string) {
	if len(users) == 0 {
		return
	}
	result := make([]map[string]any, 0, len(users))
	for _, u := range users {
		user := map[string]any{
			"name":   u.Name,
			"gecos":  u.Gecos,
			"groups": u.Groups,
			"shell":  u.Shell,
			"home":   u.Home,
			"sudo":   u.Sudo,
		}
		if u.Name == "boite" {
			seen := make(map[string]bool)
			keys := make([]string, 0, len(u.SSHAuthorizedKeys)+1)
			for _, key := range u.SSHAuthorizedKeys {
				if !seen[key] {
					seen[key] = true
					keys = append(keys, key)
				}
			}
			if !seen[sshPubKey] {
				keys = append(keys, sshPubKey)
			}
			user["ssh_authorized_keys"] = keys
		}
		result = append(result, user)
	}
	cloudConfig["users"] = result
}

func addPackagesToConfig(cloudConfig map[string]any, packages []string) {
	if len(packages) == 0 {
		return
	}
	pkgs := append([]string(nil), packages...)
	cloudConfig["packages"] = pkgs
}

func addAPTSourcesToConfig(cloudConfig map[string]any, apt *APTConfig) {
	if apt == nil || len(apt.Sources) == 0 {
		return
	}
	sources := make(map[string]any)
	for name, src := range apt.Sources {
		source := src.Source
		if strings.Contains(source, "download.docker.com/linux/ubuntu") {
			source = strings.Replace(source, "linux/ubuntu", "linux/debian", 1)
			fmt.Fprintf(os.Stderr, "Info: converted Docker APT source to Debian variant: %s\n", source)
		}
		if strings.Contains(source, "ubuntu") || strings.Contains(source, "docker.io") {
			fmt.Fprintf(os.Stderr, "Warning: skipping Ubuntu-specific APT source '%s' for Debian 12 compatibility\n", name)
			continue
		}
		sources[name] = map[string]any{
			"keyid":  src.KeyID,
			"source": source,
		}
	}
	if len(sources) > 0 {
		cloudConfig["apt"] = map[string]any{
			"sources": sources,
		}
	}
}

func addRuncmdToConfig(cloudConfig map[string]any, runcmd []any) {
	var result []any
	if len(runcmd) > 0 {
		result = make([]any, 0, len(runcmd))
		for _, r := range runcmd {
			if s, ok := r.(string); ok {
				result = append(result, s)
			} else {
				result = append(result, fmt.Sprintf("%v", r))
			}
		}
	}
	result = append([]any{"netplan set ethernets.eth0.dhcp4=false"}, result...)
	cloudConfig["runcmd"] = result
}

func addWriteFilesToConfig(cloudConfig map[string]any, writeFiles []WriteFileConfig) {
	if len(writeFiles) == 0 {
		return
	}
	result := make([]map[string]any, 0, len(writeFiles))
	for _, w := range writeFiles {
		file := map[string]any{
			"path":    w.Path,
			"content": w.Content,
		}
		if w.Owner != "" {
			file["owner"] = w.Owner
		}
		if w.Permissions != "" {
			file["permissions"] = w.Permissions
		}
		if w.Encoding != "" {
			file["encoding"] = w.Encoding
		}
		result = append(result, file)
	}
	cloudConfig["write_files"] = result
}

func writeCloudInitFiles(userDataPath, userData, metaDataPath, metaData string) error {
	if err := os.WriteFile(userDataPath, []byte(userData), 0o644); err != nil {
		return fmt.Errorf("write user-data: %w", err)
	}
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0o644); err != nil {
		return fmt.Errorf("write meta-data: %w", err)
	}
	return nil
}

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
