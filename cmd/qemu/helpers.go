package qemu

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type startFinalizeParams struct {
	name          string
	workspacePath string
	noMount       bool
	overlayPath   string
	configISOPath string
	configPath    string
	keyResolution sshKeyResolution
	cfg           *BoiteConfig
}

func (p *startFinalizeParams) buildInstance(pid int, sshPort int) *Instance {
	return &Instance{
		Name:          p.name,
		PID:           pid,
		SSHPort:       sshPort,
		OverlayPath:   p.overlayPath,
		ConfigISOPath: p.configISOPath,
		KeyPath:       p.keyResolution.privateKeyPath,
		PubKeyPath:    p.keyResolution.publicKeyPath,
		CreatedAt:     time.Now(),
		Status:        "running",
		Workspace:     p.workspacePath,
		NoMount:       p.noMount,
	}
}

func setupCleanupHandler(name string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		inst, err := LoadInstanceState(name)
		if err == nil && inst.PID > 0 {
			if err := KillQEMU(inst.PID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to stop QEMU: %v\n", err)
			}
		}
		os.Exit(130)
	}()
}

func applyVMConfig(qemuCfg *QEMUConfig, cfg *BoiteConfig) {
	if cfg != nil && cfg.VM != nil {
		qemuCfg.Memory = cfg.VM.Memory
		qemuCfg.CPUs = cfg.VM.CPUs
	}
}

func vmDiskSize(cfg *BoiteConfig) int {
	if cfg == nil || cfg.VM == nil || cfg.VM.Disk == "" {
		return 20
	}
	return parseDiskGB(cfg.VM.Disk)
}

func parseDiskGB(value string) int {
	value = strings.TrimSpace(strings.ToLower(value))
	if strings.HasSuffix(value, "g") {
		value = strings.TrimSuffix(value, "g")
	}
	if n, err := strconv.Atoi(value); err == nil && n > 0 {
		return n
	}
	return 20
}

func WaitForSSH(port, timeoutSeconds int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for i := 0; i < timeoutSeconds; i++ {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			ProgressDone(fmt.Sprintf("SSH up on %s (%ds)", addr, i+2))
			time.Sleep(2 * time.Second)
			return nil
		}
		ProgressTick(fmt.Sprintf("Waiting for SSH on %s (%ds)...", addr, i+1), float64(i)/float64(timeoutSeconds))
		time.Sleep(1 * time.Second)
	}
	ProgressFail("SSH never came up")
	return fmt.Errorf("timeout waiting for SSH on %s", addr)
}

func WaitForProcessExit(pid int, timeout int) error {
	for i := 0; i < timeout; i++ {
		if !IsProcessRunning(pid) {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for process %d to exit", pid)
}

func WaitForPortFree(port int, timeout int) error {
	for i := 0; i < timeout; i++ {
		if isPortFree(port) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for port %d to be freed", port)
}
