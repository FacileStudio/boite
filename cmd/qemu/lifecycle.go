package qemu

import (
	"fmt"
	"os"
)

// Create provisions a brand new instance: a fresh overlay disk, its SSH key,
// a config disk with the environment, then boots the VM and waits for it to
// finish firstboot. Returns the running instance, or (nil, err) on failure.
func Create(name, workspacePath string, noMount bool, configPath string, generateKey bool) (*Instance, error) {
	if InstanceExists(name) {
		return nil, fmt.Errorf("instance '%s' already exists", name)
	}

	cfg, err := LoadBoiteConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load boite config: %w", err)
	}

	ProgressPhase("Preparing instance disk")
	overlayPath, err := prepareInstanceDisk(name, cfg)
	if err != nil {
		return nil, err
	}

	configDiskPath, keyResolution, err := prepareConfig(GetInstanceDir(name), generateKey, cfg)
	if err != nil {
		return nil, err
	}

	ProgressPhase("Starting VM")
	inst, err := startAndFinalizeInstance(&startFinalizeParams{
		name: name, workspacePath: workspacePath, noMount: noMount, overlayPath: overlayPath,
		configDiskPath: configDiskPath, configPath: configPath, keyResolution: keyResolution, cfg: cfg,
	})
	if err != nil {
		return nil, err
	}

	setupCleanupHandler(name)
	return inst, nil
}

// startAndFinalizeInstance boots the prepared VM, records its state and
// waits for the guest to come up, run firstboot and accept provisioning.
func startAndFinalizeInstance(p *startFinalizeParams) (*Instance, error) {
	sshPort, err := FindFreePort()
	if err != nil {
		return nil, fmt.Errorf("find free port: %w", err)
	}

	qemuCfg := QEMUConfig{
		BaseImage: GetBaseImagePath(), OverlayPath: p.overlayPath, ConfigDiskPath: p.configDiskPath,
		HostFwdPort: sshPort + 1, PIDFile: GetPIDPath(p.name), ConsoleLog: GetConsoleLogPath(p.name),
	}
	applyVMConfig(&qemuCfg, p.cfg)

	pid, err := launchVM(qemuCfg)
	if err != nil {
		return nil, err
	}
	ProgressDone(fmt.Sprintf("QEMU running (pid %d, SSH on port %d)", pid, sshPort+1))

	inst := p.buildInstance(pid, qemuCfg.HostFwdPort)
	if err := SaveInstanceState(inst); err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	if err := WaitForSSH(inst.SSHPort, 120); err != nil {
		return nil, fmt.Errorf("wait for ssh: %w", err)
	}

	if err := WaitForFirstboot(inst, 180); err != nil {
		return nil, fmt.Errorf("firstboot: %w", err)
	}

	if err := Provision(inst, p.cfg); err != nil {
		ProgressFail("provisioning failed")
		return nil, fmt.Errorf("provision: %w", err)
	}

	return inst, nil
}

// prepareConfig resolves or generates the instance SSH key, then materialises
// the config disk from the boite config: everything the guest needs before boot.
func prepareConfig(instanceDir string, generateKey bool, cfg *BoiteConfig) (string, sshKeyResolution, error) {
	ProgressPhase("Resolving SSH key")
	keyResolution, err := resolveSSHKey(instanceDir, generateKey)
	if err != nil {
		return "", keyResolution, err
	}

	ProgressPhase("Building config disk")
	envVars, err := materializeEnv(cfg)
	if err != nil {
		return "", keyResolution, fmt.Errorf("resolve env: %w", err)
	}
	configDiskPath, err := BuildConfigDisk(instanceDir, keyResolution.pubKey, envVars)
	if err != nil {
		return configDiskPath, keyResolution, fmt.Errorf("build config disk: %w", err)
	}

	return configDiskPath, keyResolution, nil
}

// launchVM starts qemu with the given config and blocks until its PID file
// exists, so the caller can rely on the process being alive.
func launchVM(cfg QEMUConfig) (int, error) {
	if _, err := StartQEMU(cfg); err != nil {
		return 0, fmt.Errorf("start qemu: %w", err)
	}

	pid, err := WaitForPID(cfg.PIDFile, 20)
	if err != nil {
		return 0, fmt.Errorf("wait for pid file: %w", err)
	}
	return pid, nil
}

func prepareInstanceDisk(name string, cfg *BoiteConfig) (string, error) {
	baseImage, err := EnsureBaseImage()
	if err != nil {
		return "", fmt.Errorf("base image: %w", err)
	}

	instanceDir := GetInstanceDir(name)
	if err := os.MkdirAll(instanceDir, 0o755); err != nil {
		return "", fmt.Errorf("prepare instance dir: %w", err)
	}

	overlayPath := GetOverlayPath(name)
	if err := CreateOverlay(baseImage, overlayPath, vmDiskSize(cfg)); err != nil {
		return "", fmt.Errorf("create overlay: %w", err)
	}

	return overlayPath, nil
}

// Start boots an existing instance: it loads the saved state, starts qemu
// if the process is not already running, and waits for SSH to come up.
func Start(name string, configPath string) (*Instance, error) {
	inst, err := LoadInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	if inst.Status == "running" && IsProcessRunning(inst.PID) {
		return inst, nil
	}

	cfg, err := LoadBoiteConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load boite config: %w", err)
	}

	qemuCfg := QEMUConfig{
		BaseImage: GetBaseImagePath(), OverlayPath: inst.OverlayPath, ConfigDiskPath: inst.ConfigDiskPath,
		HostFwdPort: inst.SSHPort, PIDFile: GetPIDPath(name), ConsoleLog: GetConsoleLogPath(name),
	}
	applyVMConfig(&qemuCfg, cfg)

	pid, err := launchVM(qemuCfg)
	if err != nil {
		return nil, err
	}
	inst.PID, inst.Status = pid, "running"
	if err := SaveInstanceState(inst); err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	if err := WaitForSSH(inst.SSHPort, 60); err != nil {
		return nil, fmt.Errorf("wait for ssh: %w", err)
	}

	setupCleanupHandler(name)
	return inst, nil
}

// Stop powers the instance down: the guest is asked to power off cleanly so
// its filesystems flush, and only if that fails or stalls is the qemu process
// killed from the host. The instance is marked stopped either way.
func Stop(name string) error {
	inst, err := LoadInstanceState(name)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	if inst.PID <= 0 || !IsProcessRunning(inst.PID) {
		inst.Status = "stopped"
		return SaveInstanceState(inst)
	}

	if err := GracefulGuestShutdown(inst, 30); err != nil {
		if err := KillQEMU(inst.PID); err != nil {
			return fmt.Errorf("kill qemu: %w", err)
		}
	}
	if err := WaitForProcessExit(inst.PID, 10); err != nil {
		return fmt.Errorf("wait for qemu to exit: %w", err)
	}
	if err := WaitForPortFree(inst.SSHPort, 20); err != nil {
		return fmt.Errorf("wait for port to free: %w", err)
	}

	inst.Status = "stopped"
	return SaveInstanceState(inst)
}

// Destroy removes the instance's state, overlay and config disk, stopping
// the VM first if it is still running. Missing state is not an error: the
// instance is already gone.
func Destroy(name string) error {
	inst, err := LoadInstanceState(name)
	if err == nil {
		if inst.PID > 0 {
			if err := Stop(name); err != nil {
				return fmt.Errorf("stop before destroy: %w", err)
			}
		}
		return DeleteInstanceDir(name)
	}
	if os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("load state: %w", err)
}
