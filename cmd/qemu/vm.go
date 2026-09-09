package qemu

import (
	"fmt"
	"time"
)

func startQEMUInstance(cfg QEMUConfig) (*Instance, error) {
	if _, err := StartQEMU(cfg); err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	pid, err := WaitForPID(cfg.PIDFile, 20)
	if err != nil {
		return nil, fmt.Errorf("wait for pid file: %w", err)
	}

	inst := &Instance{
		PID:         pid,
		SSHPort:     cfg.HostFwdPort,
		OverlayPath: cfg.OverlayPath,
		SeedISOPath: cfg.SeedISOPath,
		CreatedAt:   time.Now(),
		Status:      "running",
	}
	return inst, nil
}

func waitForInstanceReady(inst *Instance) error {
	if err := WaitForSSH(inst.SSHPort, 120); err != nil {
		return fmt.Errorf("wait for ssh: %w", err)
	}

	if err := WaitForCloudInit(inst, 180); err != nil {
		return fmt.Errorf("wait for cloud-init: %w", err)
	}
	return nil
}
