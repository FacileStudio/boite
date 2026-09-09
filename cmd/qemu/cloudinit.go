package qemu

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
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
		return "", fmt.Errorf("no 'boite' user defined in config; add a cloud_init.users entry for 'boite' with an SSH key")
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

// WaitForCloudInit blocks until cloud-init reports it has finished on the
// instance, surfacing its status and the latest VM console line while it runs.
// It fails on "disabled" (cloud-init never provisioned) or when the guest never
// confirms cloud-init started. It never aborts a confirmed first boot with a
// heavy runcmd on a wall-clock deadline: it keeps polling until "done", and a
// boot that completes with module errors ("error") settles as a warning rather
// than failing the create.
func WaitForCloudInit(inst *Instance, timeoutSeconds int) error {
	steps := timeoutSeconds / 5
	if steps < 1 {
		steps = 1
	}
	capSteps := 900 / 5
	sawRunning := false
	elapsed := 0
	for i := 0; i < capSteps; i++ {
		detail := cloudInitStatus(inst)
		state := cloudInitState(detail)

		if state == "done" {
			ProgressDone(fmt.Sprintf("cloud-init finished (%ds)", elapsed))
			return nil
		}
		if state == "error" {
			ProgressWarn(fmt.Sprintf("cloud-init finished with errors (%ds)", elapsed))
			return nil
		}
		if isFatalCloudInitState(state) {
			ProgressFail(fmt.Sprintf("cloud-init %s", detail))
			return fmt.Errorf("cloud-init %s", detail)
		}

		if state == "running" {
			sawRunning = true
		}

		if live := cloudInitLiveLog(inst); live != "" {
			detail = fmt.Sprintf("%s · %s", detail, live)
		}
		frac := float64(i) / float64(steps)
		if frac > 1.0 {
			frac = 1.0
		}
		ProgressTick(fmt.Sprintf("cloud-init (%ds): %s", elapsed, detail), frac)
		time.Sleep(5 * time.Second)
		elapsed += 5

		if !sawRunning && elapsed >= timeoutSeconds {
			ProgressFail("cloud-init never became reachable")
			return fmt.Errorf("cloud-init: never started or never reported status within %ds", timeoutSeconds)
		}
	}
	ProgressFail("cloud-init did not finish")
	return fmt.Errorf("cloud-init: still running after %ds", elapsed)
}

// cloudInitState extracts the bare state word ("running", "done", "error", ...)
// from `cloud-init status` output, or "" when the guest is not reporting yet.
func cloudInitState(detail string) string {
	if !strings.Contains(detail, "status: ") {
		return ""
	}
	parts := strings.Split(detail, "status: ")
	state := strings.TrimSpace(parts[len(parts)-1])
	return strings.Split(state, "\n")[0]
}

// isFatalCloudInitState reports whether the state means cloud-init never
// provisioned the guest, so create should fail instead of waiting. "error"
// is deliberately excluded: it means the boot completed but a non-fatal module
// failed, and waitForCloudInit surface it as a warning. "disabled" means
// cloud-init never ran, so no user or key was ever created.
func isFatalCloudInitState(state string) bool {
	return state == "disabled" || state == "failed" || state == "canceled"
}

// cloudInitStatus runs `cloud-init status` over ssh and returns the guest's
// status line verbatim, or "not reporting yet" when the guest is unreachable.
// The remote command exits nonzero while cloud-init is still running
// (documented exit code 2), so a nonzero ssh exit is NOT a transport failure:
// judge reachability by the output content, not the exit code. Keying on the
// "status: " line means a real connection/auth error (ssh writes its refusal
// to stderr, and CombinedOutput merges stderr) yields no status line and falls
// back to "not reporting yet", while "status: running" survives a nonzero exit.
func cloudInitStatus(inst *Instance) string {
	cmd := exec.Command("ssh",
		"-i", getSSHIdentityFile(inst),
		"-p", fmt.Sprintf("%d", inst.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ConnectTimeout=10",
		"boite@127.0.0.1",
		"cloud-init status",
	)
	out, _ := cmd.CombinedOutput()
	status := strings.TrimSpace(string(out))
	if !strings.Contains(status, "status: ") {
		return "not reporting yet"
	}
	return status
}

func cloudInitLiveLog(inst *Instance) string {
	data, err := os.ReadFile(GetConsoleLogPath(inst.Name))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if len(line) > 96 {
			line = line[:96] + "…"
		}
		return line
	}
	return ""
}
