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

func Create(name, workspacePath string, noMount bool) (*Instance, error) {
	if InstanceExists(name) {
		return nil, fmt.Errorf("instance '%s' already exists", name)
	}

	baseImage, err := EnsureBaseImage()
	if err != nil {
		return nil, fmt.Errorf("base image: %w", err)
	}

	instanceDir, err := prepareInstanceDir(name)
	if err != nil {
		return nil, fmt.Errorf("prepare instance dir: %w", err)
	}

	overlayPath, err := createInstanceOverlay(baseImage, name)
	if err != nil {
		return nil, fmt.Errorf("create overlay: %w", err)
	}

	cfg, _ := LoadBoiteConfig()
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

	seedISOPath, err := GenerateSeedISO(instanceDir, name, pubKey)
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

	pid, err := startQEMUAndWait(startQEMUParams{
		BaseImage:   baseImage,
		OverlayPath: overlayPath,
		SeedISOPath: seedISOPath,
		HostFwdPort: hostfwdPort,
		PIDFile:     pidFile,
		ConsoleLog:  consoleLog,
	})
	if err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	inst := newInstance(newInstanceParams{
		Name:           name,
		PID:            pid,
		HostFwdPort:    hostfwdPort,
		OverlayPath:    overlayPath,
		SeedISOPath:    seedISOPath,
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		WorkspacePath:  workspacePath,
		NoMount:        noMount,
	})
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

func waitForPIDFile(pidFile string, timeout int) error {
	for range timeout {
		if _, err := os.Stat(pidFile); err == nil {
			data, _ := os.ReadFile(pidFile)
			if len(data) > 0 {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("PID file not created")
}

func WaitForSSH(port, timeoutSeconds int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for range timeoutSeconds {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			if err := conn.Close(); err != nil {
				return fmt.Errorf("close connection: %w", err)
			}
			time.Sleep(2 * time.Second)
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for SSH on %s", addr)
}

// WaitForCloudInit waits for cloud-init to finish running after SSH is reachable.
// Uses polling instead of --wait to avoid blocking.
func WaitForCloudInit(inst *Instance, timeoutSeconds int) error {
	consoleLog := GetConsoleLogPath(inst.Name)
	for range timeoutSeconds {
		if data, err := os.ReadFile(consoleLog); err == nil {
			if strings.Contains(string(data), "Cloud-init v.") && strings.Contains(string(data), "Finished at") {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for cloud-init to complete")
}

func Start(name string) (*Instance, error) {
	inst, err := LoadInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	if inst.Status == "running" && IsProcessRunning(inst.PID) {
		return inst, nil
	}

	pidFile := GetPIDPath(name)
	consoleLog := GetConsoleLogPath(inst.Name)
	baseImage := GetBaseImagePath()

	cfg := QEMUConfig{BaseImage: baseImage, OverlayPath: inst.OverlayPath, SeedISOPath: inst.SeedISOPath, HostFwdPort: inst.SSHPort, PIDFile: pidFile, ConsoleLog: consoleLog}
	cmdQEMU, err := StartQEMU(cfg)
	if err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	pid := cmdQEMU.Process.Pid
	if err := waitForPIDFile(pidFile, 10); err != nil {
		return nil, fmt.Errorf("wait for pid file: %w", err)
	}

	pidData, err := os.ReadFile(pidFile)
	if err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		if pidInt, err2 := strconv.Atoi(pidStr); err2 == nil && pidInt > 0 {
			pid = pidInt
		}
	}

	inst.PID = pid
	inst.Status = "running"
	if err := SaveInstanceState(inst); err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	if err := WaitForSSH(inst.SSHPort, 60); err != nil {
		return nil, fmt.Errorf("wait for ssh: %w", err)
	}

	if err := WaitForCloudInit(inst, 120); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cloud-init may not have completed: %v\n", err)
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
		time.Sleep(1 * time.Second)
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
		time.Sleep(1 * time.Second)
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
			if err := KillQEMU(inst.PID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to kill qemu: %v\n", err)
			}
		}
		os.Exit(130)
	}()
}

