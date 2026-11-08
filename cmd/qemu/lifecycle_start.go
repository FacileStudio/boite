package qemu

import "fmt"

// StartDetached boots an existing instance: it loads the saved state, starts
// qemu if the process is not already running, and waits for SSH to come up.
// Unlike Start it installs no signal handler, so a daemon's boxes outlive it.
func StartDetached(name string, configPath string) (*Instance, error) {
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
		BaseImage: GetBaseImagePath(inst.Distro), OverlayPath: inst.OverlayPath, ConfigDiskPath: inst.ConfigDiskPath,
		HostFwdPort: inst.SSHPort, PIDFile: GetPIDPath(name), ConsoleLog: GetConsoleLogPath(name),
	}
	applyVMConfig(&qemuCfg, cfg, GetInstanceDir(name))

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

	return inst, nil
}

// Start boots an existing instance like StartDetached, then installs the
// handler that kills qemu when the CLI is interrupted.
func Start(name string, configPath string) (*Instance, error) {
	inst, err := StartDetached(name, configPath)
	if err != nil {
		return nil, err
	}
	setupCleanupHandler(name)
	return inst, nil
}
