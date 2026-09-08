package qemu

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	keygen "github.com/charmbracelet/keygen"
	"gopkg.in/yaml.v3"
)

// GenerateSSHKeyPair creates an Ed25519 key pair using keygen and writes
// them to the central key store at ~/.boite/ssh/. Existing keys are
// reused if found on disk (keygen's LoadOrCreate pattern).
func GenerateSSHKeyPair(instanceDir string) (string, string, error) {
	storeDir := filepath.Join(GetBoiteDir(), "ssh")
	if err := os.MkdirAll(storeDir, 0o700); err != nil {
		return "", "", fmt.Errorf("create key store dir: %w", err)
	}

	keyPath := filepath.Join(storeDir, "id_ed25519")
	kp, err := keygen.New(keyPath, keygen.WithKeyType(keygen.Ed25519), keygen.WithWrite())
	if err != nil {
		return "", "", fmt.Errorf("keygen: %w", err)
	}

	privatePath := keyPath
	publicPath := keyPath + ".pub"

	privBytes := kp.RawPrivateKey()
	if err := os.WriteFile(privatePath, privBytes, 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
	}
	if err := os.Chmod(privatePath, 0o600); err != nil {
		return "", "", fmt.Errorf("chmod private key: %w", err)
	}

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

// writeCloudInitFiles writes user-data and meta-data to the instance directory
func writeCloudInitFiles(userDataPath, userData, metaDataPath, metaData string) error {
	if err := os.WriteFile(userDataPath, []byte(userData), 0o644); err != nil {
		return fmt.Errorf("write user-data: %w", err)
	}
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0o644); err != nil {
		return fmt.Errorf("write meta-data: %w", err)
	}
	return nil
}

// runCloudLocalDS creates a seed ISO using mkisofs/genisoimage with NoCloud datasource layout
func runCloudLocalDS(seedISOPath, userDataPath, metaDataPath string) error {
	mkisofs, err := exec.LookPath("mkisofs")
	if err != nil {
		mkisofs, err = exec.LookPath("genisoimage")
		if err != nil {
			return fmt.Errorf("mkisofs/genisoimage not found: %w", err)
		}
	}

	args := []string{
		"-output", seedISOPath,
		"-volid", "cidata",
		"-joliet",
		"-rock",
		userDataPath,
		metaDataPath,
	}

	cmd := exec.Command(mkisofs, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("genisoimage failed: %s (%w)", string(output), err)
	}
	return nil
}
func GenerateSeedISO(instanceDir, instanceName, sshPubKey string) (string, error) {
	cfg, err := LoadBoiteConfig()
	if err != nil {
		return "", fmt.Errorf("load boite config: %w", err)
	}

	mergedCfg := MergeCloudInitConfig(cfg)
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

	cloudConfig := buildCloudConfig(cfg, sshPubKey)

	encoder.Encode(cloudConfig)
	encoder.Close()

	return "#cloud-config\n" + buf.String()
}

// buildCloudConfig constructs the cloud-init configuration map
func buildCloudConfig(cfg *CloudInitConfig, sshPubKey string) map[string]interface{} {
	cloudConfig := map[string]interface{}{
		"hostname":       "boite",
		"disable_root":   true,
		"ssh_pwauth":     false,
		"package_update": true,
		"network":        buildNetworkConfig(),
	}

	injectUsers(cloudConfig, cfg, sshPubKey)
	injectPackages(cloudConfig, cfg)
	injectAPTSources(cloudConfig, cfg)
	injectRuncmd(cloudConfig, cfg)
	injectWriteFiles(cloudConfig, cfg)

	return cloudConfig
}

// buildNetworkConfig returns the network configuration
func buildNetworkConfig() map[string]interface{} {
	return map[string]interface{}{
		"version": 2,
		"ethernets": map[string]interface{}{
			"eth0": map[string]interface{}{
				"dhcp4":     false,
				"addresses": []string{"192.168.42.10/24"},
				"routes":    []map[string]interface{}{{"to": "default", "via": "192.168.42.1"}},
				"nameservers": map[string]interface{}{
					"addresses": []string{"8.8.8.8", "8.8.4.4"},
				},
			},
		},
	}
}

// injectUsers adds users to the cloud config if configured
func injectUsers(cloudConfig map[string]interface{}, cfg *CloudInitConfig, sshPubKey string) {
	if len(cfg.Users) > 0 {
		users := make([]map[string]interface{}, 0, len(cfg.Users))
		for _, u := range cfg.Users {
			user := buildUserConfig(u, sshPubKey)
			users = append(users, user)
		}
		cloudConfig["users"] = users
	}
}

// buildUserConfig builds a single user configuration map
func buildUserConfig(u UserConfig, sshPubKey string) map[string]interface{} {
	user := map[string]interface{}{
		"name":   u.Name,
		"gecos":  u.Gecos,
		"groups": u.Groups,
		"shell":  u.Shell,
		"home":   u.Home,
		"sudo":   u.Sudo,
	}
	if u.Name == "boite" {
		keys := make([]string, 0, len(u.SSHAuthorizedKeys)+1)
		keys = append(keys, u.SSHAuthorizedKeys...)
		keys = append(keys, sshPubKey)
		user["ssh_authorized_keys"] = keys
	}
	return user
}

// injectPackages adds packages to the cloud config if configured
func injectPackages(cloudConfig map[string]interface{}, cfg *CloudInitConfig) {
	if len(cfg.Packages) > 0 {
		pkgs := make([]string, 0, len(cfg.Packages))
		for _, p := range cfg.Packages {
			pkgs = append(pkgs, p)
		}
		cloudConfig["packages"] = pkgs
	}
}

// injectAPTSources adds APT sources to the cloud config if configured
func injectAPTSources(cloudConfig map[string]interface{}, cfg *CloudInitConfig) {
	if cfg.APT != nil && len(cfg.APT.Sources) > 0 {
		sources := make(map[string]interface{})
		for name, src := range cfg.APT.Sources {
			if !strings.Contains(src.Source, "ubuntu") && !strings.Contains(src.Source, "docker.io") {
				sources[name] = map[string]interface{}{
					"keyid":  src.KeyID,
					"source": src.Source,
				}
			}
		}
		if len(sources) > 0 {
			cloudConfig["apt"] = map[string]interface{}{
				"sources": sources,
			}
		}
	}
}

// injectRuncmd adds run commands to the cloud config
func injectRuncmd(cloudConfig map[string]interface{}, cfg *CloudInitConfig) {
	if len(cfg.Runcmd) > 0 {
		runcmd := make([]interface{}, 0, len(cfg.Runcmd))
		runcmd = append(runcmd, cfg.Runcmd...)
		cloudConfig["runcmd"] = runcmd
	}

	if len(cfg.Runcmd) == 0 {
		cloudConfig["runcmd"] = []interface{}{"netplan set ethernets.eth0.dhcp4=false"}
	}
}

// injectWriteFiles adds write_files to the cloud config if configured
func injectWriteFiles(cloudConfig map[string]interface{}, cfg *CloudInitConfig) {
	if len(cfg.WriteFiles) > 0 {
		files := make([]map[string]interface{}, 0, len(cfg.WriteFiles))
		for _, f := range cfg.WriteFiles {
			file := map[string]interface{}{
				"path": f.Path,
			}
			if f.Permissions != "" {
				file["permissions"] = f.Permissions
			}
			if f.Owner != "" {
				file["owner"] = f.Owner
			}
			if f.Content != "" {
				file["content"] = f.Content
			}
			files = append(files, file)
		}
		cloudConfig["write_files"] = files
	}
}

// ComputeInstanceID generates a short instance ID from the name
func ComputeInstanceID(name string) string {
	h := sha256.Sum256([]byte(name + "boite-instance"))
	return fmt.Sprintf("%x", h[:8])
}
