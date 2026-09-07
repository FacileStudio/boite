package qemu

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

// GenerateSSHKeyPair creates an ed25519 key pair and writes them to disk
func GenerateSSHKeyPair(instanceDir string) (string, string, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("generate key: %w", err)
	}

	privatePath := filepath.Join(instanceDir, "id_ed25519")
	publicPath := filepath.Join(instanceDir, "id_ed25519.pub")

	privatePEM, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return "", "", fmt.Errorf("marshal private key: %w", err)
	}
	if err := os.WriteFile(privatePath, pem.EncodeToMemory(privatePEM), 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
	}

	pubKey, err := ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		return "", "", fmt.Errorf("ssh public key: %w", err)
	}
	pubBytes := ssh.MarshalAuthorizedKey(pubKey)
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

// GenerateSeedISO creates the seed.iso from the merged cloud-init config
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

	cloudConfig := map[string]interface{}{
		"hostname":      "boite",
		"disable_root":  true,
		"ssh_pwauth":    false,
		"package_update": true,
		"network": map[string]interface{}{
			"version": 2,
			"ethernets": map[string]interface{}{
				"eth0": map[string]interface{}{
					"dhcp4": false,
					"addresses": []string{
						"192.168.42.10/24",
					},
					"routes": []map[string]interface{}{
						{"to": "default", "via": "192.168.42.1"},
					},
					"nameservers": map[string]interface{}{
						"addresses": []string{"8.8.8.8", "8.8.4.4"},
					},
				},
			},
		},
	}

	// Users
	if len(cfg.Users) > 0 {
		users := make([]map[string]interface{}, 0, len(cfg.Users))
		for _, u := range cfg.Users {
			user := map[string]interface{}{
				"name":    u.Name,
				"gecos":   u.Gecos,
				"groups":  u.Groups,
				"shell":   u.Shell,
				"home":    u.Home,
				"sudo":    u.Sudo,
			}
			// Always inject the generated instance key into the boite user's
			// authorized_keys, merging with any keys the user configured.
			if u.Name == "boite" {
				keys := make([]string, 0, len(u.SSHAuthorizedKeys)+1)
				keys = append(keys, u.SSHAuthorizedKeys...)
				keys = append(keys, sshPubKey)
				user["ssh_authorized_keys"] = keys
			}
			users = append(users, user)
		}
		cloudConfig["users"] = users
	}

	// Packages
	if len(cfg.Packages) > 0 {
		pkgs := make([]string, 0, len(cfg.Packages))
		for _, p := range cfg.Packages {
			pkgs = append(pkgs, p)
		}
		cloudConfig["packages"] = pkgs
	}

	// APT sources
	if cfg.APT != nil && len(cfg.APT.Sources) > 0 {
		sources := make(map[string]interface{})
		for name, src := range cfg.APT.Sources {
			// Skip Ubuntu-specific APT sources for Debian 12 compatibility
			if strings.Contains(src.Source, "ubuntu") || strings.Contains(src.Source, "docker.io") {
				fmt.Fprintf(os.Stderr, "Warning: skipping Ubuntu-specific APT source '%s' for Debian 12 compatibility\n", name)
				continue
			}
			sources[name] = map[string]interface{}{
				"keyid":   src.KeyID,
				"source":  src.Source,
			}
		}
		if len(sources) > 0 {
			cloudConfig["apt"] = map[string]interface{}{
				"sources": sources,
			}
		}
	}

	// runcmd
	if len(cfg.Runcmd) > 0 {
		runcmd := make([]interface{}, 0, len(cfg.Runcmd))
		for _, r := range cfg.Runcmd {
			runcmd = append(runcmd, r)
		}
		cloudConfig["runcmd"] = runcmd
	}

	// Add netplan set ethernets.eth0.dhcp4=false for static IP configuration
	cloudConfig["runcmd"] = append(
		[]interface{}{"netplan set ethernets.eth0.dhcp4=false"},
		cloudConfig["runcmd"].([]interface{})...,
	)

	encoder.Encode(cloudConfig)
	encoder.Close()

	return "#cloud-config\n" + buf.String()
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