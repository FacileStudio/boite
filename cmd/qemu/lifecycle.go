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

func Create(name, workspacePath string, noMount bool) (*Instance, error) {
	if InstanceExists(name) {
		return nil, fmt.Errorf("instance '%s' already exists", name)
	}

	baseImage, err := EnsureBaseImage()
	if err != nil {
		return nil, fmt.Errorf("base image: %w", err)
	}

	instanceDir := GetInstanceDir(name)
	if err := os.MkdirAll(instanceDir, 0o755); err != nil {
		return nil, fmt.Errorf("create instance dir: %w", err)
	}

	overlayPath := GetOverlayPath(name)
	if err := CreateOverlay(baseImage, overlayPath, 10); err != nil {
		return nil, fmt.Errorf("create overlay: %w", err)
	}

	privateKeyPath, publicKeyPath, err := GenerateSSHKeyPair(instanceDir)
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
	cmdQEMU, err := StartQEMU(baseImage, overlayPath, seedISOPath, sshPort, pidFile)
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

	inst := &Instance{
		Name:         name,
		PID:          pid,
		SSHPort:      hostfwdPort,
		OverlayPath:  overlayPath,
		SeedISOPath:  seedISOPath,
		KeyPath:      privateKeyPath,
		PubKeyPath:   publicKeyPath,
		CreatedAt:    time.Now(),
		Status:       "running",
		Workspace:    workspacePath,
		NoMount:      noMount,
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

func waitForPIDFile(pidFile string, timeout int) error {
	for i := 0; i < timeout; i++ {
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

// WaitForCloudInit waits for cloud-init to finish running after SSH is reachable.
// Uses polling instead of --wait to avoid blocking.
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

func Start(name string) (*Instance, error) {
	inst, err := LoadInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	if inst.Status == "running" && IsProcessRunning(inst.PID) {
		return inst, nil
	}

	pidFile := GetPIDPath(name)
	baseImage := GetBaseImagePath()

	cmdQEMU, err := StartQEMU(baseImage, inst.OverlayPath, inst.SeedISOPath, inst.SSHPort, pidFile)
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

	setupCleanupHandler(name)
	return inst, nil
}

func Stop(name string) error {
	inst, err := LoadInstanceState(name)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	if inst.PID > 0 && IsProcessRunning(inst.PID) {
		KillQEMU(inst.PID)
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
		KillQEMU(inst.PID)
		time.Sleep(1 * time.Second)
	}

	return DeleteInstanceDir(name)
}

func PurgeAll() error {
	instances, err := ListInstances()
	if err != nil {
		return err
	}
	for _, inst := range instances {
		Destroy(inst.Name)
	}
	return nil
}

func setupCleanupHandler(name string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		inst, _ := LoadInstanceState(name)
		if inst != nil && inst.PID > 0 {
			KillQEMU(inst.PID)
		}
		os.Exit(130)
	}()
}