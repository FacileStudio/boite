package qemu

import (
	"fmt"
	"os"
)

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

	instanceDir := GetInstanceDir(name)
	ProgressPhase("Resolving SSH key")
	keyResolution, err := resolveSSHKey(instanceDir, generateKey, cfg)
	if err != nil {
		return nil, err
	}

	ProgressPhase("Building config ISO")
	configISOPath, err := BuildConfigISO(instanceDir, keyResolution.pubKey)
	if err != nil {
		return nil, fmt.Errorf("build config iso: %w", err)
	}

	ProgressPhase("Starting VM")
	inst, err := startAndFinalizeInstance(&startFinalizeParams{
		name:          name,
		workspacePath: workspacePath,
		noMount:       noMount,
		overlayPath:   overlayPath,
		configISOPath: configISOPath,
		configPath:    configPath,
		keyResolution: keyResolution,
		cfg:           cfg,
	})
	if err != nil {
		return nil, err
	}

	setupCleanupHandler(name)
	return inst, nil
}

func startAndFinalizeInstance(p *startFinalizeParams) (*Instance, error) {
	sshPort, err := FindFreePort()
	if err != nil {
		return nil, fmt.Errorf("find free port: %w", err)
	}

	qemuCfg := QEMUConfig{
		BaseImage:     GetBaseImagePath(),
		OverlayPath:   p.overlayPath,
		ConfigISOPath: p.configISOPath,
		HostFwdPort:   sshPort + 1,
		PIDFile:       GetPIDPath(p.name),
		ConsoleLog:    GetConsoleLogPath(p.name),
	}
	applyVMConfig(&qemuCfg, p.cfg)

	if _, err := StartQEMU(qemuCfg); err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	pid, err := WaitForPID(qemuCfg.PIDFile, 20)
	if err != nil {
		return nil, fmt.Errorf("wait for pid file: %w", err)
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

	return inst, nil
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
		BaseImage:     GetBaseImagePath(),
		OverlayPath:   inst.OverlayPath,
		ConfigISOPath: inst.ConfigISOPath,
		HostFwdPort:   inst.SSHPort,
		PIDFile:       GetPIDPath(name),
		ConsoleLog:    GetConsoleLogPath(name),
	}
	applyVMConfig(&qemuCfg, cfg)

	if _, err := StartQEMU(qemuCfg); err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	pid, err := WaitForPID(qemuCfg.PIDFile, 20)
	if err != nil {
		return nil, fmt.Errorf("wait for pid file: %w", err)
	}

	inst.PID = pid
	inst.Status = "running"
	if err := SaveInstanceState(inst); err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	if err := WaitForSSH(inst.SSHPort, 60); err != nil {
		return nil, fmt.Errorf("wait for ssh: %w", err)
	}

	setupCleanupHandler(name)
	return inst, nil
}

func Stop(name string) error {
	inst, err := LoadInstanceState(name)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	if inst.PID > 0 && IsProcessRunning(inst.PID) {
		if err := KillQEMU(inst.PID); err != nil {
			return fmt.Errorf("kill qemu: %w", err)
		}
		if err := WaitForProcessExit(inst.PID, 10); err != nil {
			return fmt.Errorf("wait for qemu to exit: %w", err)
		}
		if err := WaitForPortFree(inst.SSHPort, 20); err != nil {
			return fmt.Errorf("wait for port to free: %w", err)
		}
	}

	inst.Status = "stopped"
	return SaveInstanceState(inst)
}

func Destroy(name string) error {
	inst, err := LoadInstanceState(name)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("load state: %w", err)
		}
		return nil
	}

	if inst.PID > 0 && IsProcessRunning(inst.PID) {
		if err := KillQEMU(inst.PID); err != nil {
			return fmt.Errorf("kill qemu: %w", err)
		}
		if err := WaitForProcessExit(inst.PID, 10); err != nil {
			return fmt.Errorf("wait for qemu to exit: %w", err)
		}
		WaitForPortFree(inst.SSHPort, 20)
	}

	return DeleteInstanceDir(name)
}
