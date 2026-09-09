package qemu

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

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

// buildBaseCloudConfig returns the base cloud-init configuration
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

// makeUserConfig creates a user configuration map for cloud-init
func makeUserConfig(u UserConfig, extraSSHKey string) map[string]any {
	user := map[string]any{
		"name":   u.Name,
		"gecos":  u.Gecos,
		"groups": u.Groups,
		"shell":  u.Shell,
		"home":   u.Home,
		"sudo":   u.Sudo,
	}
	if u.Name == "boite" {
		user["ssh_authorized_keys"] = deduplicateSSHKeys(u.SSHAuthorizedKeys, extraSSHKey)
	}
	if u.Password != "" {
		user["password"] = u.Password
	}
	return user
}

// addUsersToConfig adds users to the cloud-init configuration
func addUsersToConfig(cloudConfig map[string]any, users []UserConfig, sshPubKey string) {
	if len(users) == 0 {
		return
	}
	extraSSHKey := ""
	if sshPubKey != "" {
		for _, u := range users {
			if u.Name == "boite" && !slices.Contains(u.SSHAuthorizedKeys, sshPubKey) {
				extraSSHKey = sshPubKey
				break
			}
		}
	}

	result := make([]map[string]any, 0, len(users))
	for _, u := range users {
		user := makeUserConfig(u, extraSSHKey)
		result = append(result, user)
	}
	cloudConfig["users"] = result
}

// addPackagesToConfig adds packages to the cloud-init configuration
func addPackagesToConfig(cloudConfig map[string]any, packages []string) {
	if len(packages) == 0 {
		return
	}
	pkgs := append([]string(nil), packages...)
	cloudConfig["packages"] = pkgs
}

// addAPTSourcesToConfig adds APT sources to the cloud-init configuration
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

// addRuncmdToConfig adds runcmd to the cloud-init configuration
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
	cloudConfig["runcmd"] = result

	anyPassword := anyPasswordInUsers(cloudConfig)
	if anyPassword {
		cloudConfig["chpasswd"] = map[string]any{
			"list":   []string{"boite:boite"},
			"expire": false,
		}
		cloudConfig["ssh_pwauth"] = true
	}
}

// addWriteFilesToConfig adds write_files to the cloud-init configuration
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
