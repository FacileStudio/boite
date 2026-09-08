package qemu

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Create(name, workspacePath string, noMount bool, configPath string) (*Instance, error) {
	if InstanceExists(name) {
		return nil, fmt.Errorf("instance '%s' already exists", name)
	}

	cfg, err := LoadBoiteConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load boite config: %w", err)
	}

	baseImage, err := EnsureBaseImage()
	if err != nil {
		return nil, fmt.Errorf("base image: %w", err)
	}

	instanceDir := GetInstanceDir(name)
	if err := os.MkdirAll(instanceDir, 0o755); err != nil {
		return nil, fmt.Errorf("prepare instance dir: %w", err)
	}

	diskSize := 10
	if cfg != nil && cfg.VM != nil && cfg.VM.Disk != "" {
		diskSize = parseDiskGB(cfg.VM.Disk)
	}
	overlayPath := GetOverlayPath(name)
	if err := CreateOverlay(baseImage, overlayPath, diskSize); err != nil {
		return nil, fmt.Errorf("create overlay: %w", err)
	}

	var passphrase string
	if cfg != nil && cfg.VM != nil {
		passphrase = cfg.VM.SSHKeyPassphrase
	}
	privateKeyPath, publicKeyPath, err := GenerateSSHKeyPair(instanceDir, passphrase)
	if err != nil {
		return nil, fmt.Errorf("generate ssh keys: %w", err)
	}

	pubKey, err := ReadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	seedISOPath, err := GenerateSeedISO(instanceDir, name, pubKey, configPath)
	if err != nil {
		return nil, fmt.Errorf("generate seed iso: %w", err)
	}

	sshPort, err := FindFreePort()
	if err != nil {
		return nil, fmt.Errorf("find free port: %w", err)
	}

	hostfwdPort := sshPort + 1
	pidFile := GetPIDPath(name)
	consoleLog := GetConsoleLogPath(name)

	qemuCfg := QEMUConfig{
		BaseImage:   baseImage,
		OverlayPath: overlayPath,
		SeedISOPath: seedISOPath,
		HostFwdPort: hostfwdPort,
		PIDFile:     pidFile,
		ConsoleLog:  consoleLog,
	}
	if cfg != nil && cfg.VM != nil {
		qemuCfg.Memory = cfg.VM.Memory
		qemuCfg.CPUs = cfg.VM.CPUs
	}
	if _, err := StartQEMU(qemuCfg); err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	pid, err := WaitForPID(pidFile, 20)
	if err != nil {
		return nil, fmt.Errorf("wait for pid file: %w", err)
	}

	inst := &Instance{
		Name:        name,
		PID:         pid,
		SSHPort:     hostfwdPort,
		OverlayPath: overlayPath,
		SeedISOPath: seedISOPath,
		KeyPath:     privateKeyPath,
		PubKeyPath:  publicKeyPath,
		CreatedAt:   time.Now(),
		Status:      "running",
		Workspace:   workspacePath,
		NoMount:     noMount,
	}
	if err := SaveInstanceState(inst); err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	if err := WaitForSSH(hostfwdPort, 120); err != nil {
		return nil, fmt.Errorf("wait for ssh: %w", err)
	}

	if err := WaitForCloudInit(inst, 180); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cloud-init may not have completed: %v\n", err)
	}

	setupCleanupHandler(name)
	return inst, nil
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

	pidFile := GetPIDPath(name)
	baseImage := GetBaseImagePath()
	consoleLog := GetConsoleLogPath(name)

	qemuCfg := QEMUConfig{
		BaseImage:   baseImage,
		OverlayPath: inst.OverlayPath,
		SeedISOPath: inst.SeedISOPath,
		HostFwdPort: inst.SSHPort,
		PIDFile:     pidFile,
		ConsoleLog:  consoleLog,
	}
	if cfg != nil && cfg.VM != nil {
		qemuCfg.Memory = cfg.VM.Memory
		qemuCfg.CPUs = cfg.VM.CPUs
	}
	if _, err := StartQEMU(qemuCfg); err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	pid, err := WaitForPID(pidFile, 20)
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
		if os.IsNotExist(err) {
			return nil
		}
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

	return DeleteInstanceDir(name)
}

func setupCleanupHandler(name string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		inst, _ := LoadInstanceState(name)
		if inst != nil && inst.PID > 0 {
			_ = KillQEMU(inst.PID)
		}
		os.Exit(130)
	}()
}



func WaitForSSH(port, timeoutSeconds int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for i := 0; i < timeoutSeconds; i++ {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			time.Sleep(2 * time.Second)
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for SSH on %s", addr)
}

func WaitForCloudInit(inst *Instance, timeoutSeconds int) error {
	for i := 0; i < timeoutSeconds/5; i++ {
		cmd := exec.Command("ssh",
			"-i", inst.KeyPath,
			"-p", fmt.Sprintf("%d", inst.SSHPort),
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
			"-o", "ConnectTimeout=10",
			"boite@127.0.0.1",
			"cloud-init status",
		)
		out, err := cmd.CombinedOutput()
		if err == nil && strings.Contains(string(out), "status: done") {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timeout waiting for cloud-init to complete")
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

func parseDiskGB(value string) int {
	value = strings.TrimSpace(strings.ToLower(value))
	if strings.HasSuffix(value, "g") {
		value = strings.TrimSuffix(value, "g")
	}
	if n, err := strconv.Atoi(value); err == nil && n > 0 {
		return n
	}
	return 10
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
