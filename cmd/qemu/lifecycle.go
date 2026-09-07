package qemu

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func Create(name, workspacePath string, noMount bool) (*Instance, error) {
	preflight := []string{"qemu-system-x86_64", "qemu-img"}
	for _, bin := range preflight {
		if _, err := exec.LookPath(bin); err != nil {
			return nil, fmt.Errorf("%s not found in PATH: %w", bin, err)
		}
	}
	if _, err := exec.LookPath("mkisofs"); err != nil {
		if _, err := exec.LookPath("genisoimage"); err != nil {
			return nil, fmt.Errorf("mkisofs/genisoimage not found in PATH: %w", err)
		}
	}

	if InstanceExists(name) {
		return nil, fmt.Errorf("instance '%s' already exists", name)
	}

	baseImage, err := EnsureBaseImage()
	if err != nil {
		return nil, fmt.Errorf("base image: %w", err)
	}

	inst, err := createInstance(name, workspacePath, noMount, baseImage)
	if err != nil {
		return nil, err
	}

	if err := WaitForSSH(inst.SSHPort, 120); err != nil {
		return nil, fmt.Errorf("wait for ssh: %w", err)
	}

	if err := WaitForCloudInit(inst, 180); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cloud-init may not have completed: %v\n", err)
	}

	setupCleanupHandler(inst.Name)
	return inst, nil
}

func createInstance(name, workspacePath string, noMount bool, baseImage string) (*Instance, error) {
	cfg, err := LoadBoiteConfig()
	if err != nil {
		return nil, fmt.Errorf("load boite config: %w", err)
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
	cmdQEMU, err := StartQEMU(baseImage, overlayPath, seedISOPath, sshPort, pidFile, cfg.VM)
	if err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	pid := readPID(pidFile, cmdQEMU.Process.Pid)

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

	return inst, nil
}

func readPID(pidFile string, fallback int) int {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fallback
	}

	pidStr := strings.TrimSpace(string(data))
	pidInt, err := strconv.Atoi(pidStr)
	if err != nil || pidInt <= 0 {
		return fallback
	}

	return pidInt
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
	connected := false
	for i := 0; i < timeoutSeconds; i++ {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			if !connected {
				connected = true
				continue
			}
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

	cfg, err := LoadBoiteConfig()
	if err != nil {
		return nil, fmt.Errorf("load boite config: %w", err)
	}

	pidFile := GetPIDPath(name)
	baseImage := GetBaseImagePath()

	cmdQEMU, err := StartQEMU(baseImage, inst.OverlayPath, inst.SeedISOPath, inst.SSHPort, pidFile, cfg.VM)
	if err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}

	inst.PID = readPID(pidFile, cmdQEMU.Process.Pid)
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
			return DeleteInstanceDir(name)
		}
		return fmt.Errorf("load state: %w", err)
	}

	if inst.PID > 0 && IsProcessRunning(inst.PID) {
		KillQEMU(inst.PID)
		time.Sleep(1 * time.Second)
	}

	inst.Status = "stopped"
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

var (
	cleanupOnce sync.Once
	cleanupPIDs []int
	cleanupMu   sync.Mutex
)

func setupCleanupHandler(name string) {
	cleanupOnce.Do(func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cleanupMu.Lock()
			pids := append([]int(nil), cleanupPIDs...)
			cleanupMu.Unlock()
			for _, pid := range pids {
				KillQEMU(pid)
			}
			os.Exit(130)
		}()
	})

	inst, _ := LoadInstanceState(name)
	if inst != nil && inst.PID > 0 {
		cleanupMu.Lock()
		cleanupPIDs = append(cleanupPIDs, inst.PID)
		cleanupMu.Unlock()
	}
}