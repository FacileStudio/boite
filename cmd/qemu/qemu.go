package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func detectAcceleration() (string, []string, string) {
	switch runtime.GOOS {
	case "linux":
		if f, err := os.Open("/dev/kvm"); err == nil {
			defer f.Close()
			return "-enable-kvm", nil, "host"
		}
		return "", []string{"-machine", "accel=tcg"}, "qemu64"
	case "darwin":
		return "", []string{"-machine", "accel=hvf"}, "qemu64"
	default:
		return "", []string{"-machine", "accel=tcg"}, "qemu64"
	}
}

type QEMUConfig struct {
	BaseImage      string
	OverlayPath    string
	ConfigDiskPath string
	HostFwdPort    int
	PIDFile        string
	ConsoleLog     string
	Memory         string
	CPUs           int
}

func BuildQEMUArgs(cfg QEMUConfig) []string {
	accelFlag, accelExtra, cpuModel := detectAcceleration()
	args := []string{}
	if accelFlag != "" {
		args = append(args, accelFlag)
	}
	args = append(args, accelExtra...)

	memory := cfg.Memory
	if memory == "" {
		memory = "2G"
	}
	cpus := cfg.CPUs
	if cpus <= 0 {
		cpus = 2
	}

	args = append(args,
		"-cpu", cpuModel,
		"-m", memory,
		"-smp", strconv.Itoa(cpus),
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", cfg.OverlayPath),
		"-drive", fmt.Sprintf("file=%s,format=raw,if=virtio,readonly=on", cfg.ConfigDiskPath),
		"-netdev", fmt.Sprintf("user,id=net0,restrict=off,hostfwd=tcp:127.0.0.1:%d-:22", cfg.HostFwdPort),
		"-device", "virtio-net-pci,netdev=net0",
		"-serial", fmt.Sprintf("file:%s", cfg.ConsoleLog),
		"-display", "none",
		"-daemonize",
		"-pidfile", cfg.PIDFile,
		"-name", "boite-vm",
	)
	return args
}

func StartQEMU(cfg QEMUConfig) (*exec.Cmd, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.ConsoleLog), 0o755); err != nil {
		return nil, fmt.Errorf("create console log dir: %w", err)
	}
	f, err := os.Create(cfg.ConsoleLog)
	if err != nil {
		return nil, fmt.Errorf("create console log file: %w", err)
	}
	defer f.Close()

	args := BuildQEMUArgs(cfg)
	cmd := exec.Command("qemu-system-x86_64", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start qemu: %w", err)
	}
	return cmd, nil
}

func WaitForPID(pidFile string, timeout int) (int, error) {
	for i := 0; i < timeout; i++ {
		data, err := os.ReadFile(pidFile)
		if err == nil {
			pidStr := string(data)
			pid, err := strconv.Atoi(strings.TrimSpace(pidStr))
			if err == nil && pid > 0 {
				return pid, nil
			}
		}
		ProgressTick(fmt.Sprintf("Waiting for QEMU to start (%ds)...", i+1), float64(i)/float64(timeout))
		time.Sleep(time.Second)
	}
	return 0, fmt.Errorf("timeout waiting for PID file")
}

// GracefulGuestShutdown asks the guest OS to power itself down cleanly over
// its SSH connection (flushing guest filesystems), then waits for the qemu
// process to exit. This is the "shut it down" step: without it the qemu
// process is killed out from under the guest and buffered guest writes can be
// lost. Falls back to KillQEMU when the guest is unreachable or does not come
// down in time.
func GracefulGuestShutdown(inst *Instance, timeoutSeconds int) error {
	// Ask the guest to shut down. The boite user has NOPASSWD sudo.
	cmd := exec.Command("ssh", BuildSSHArgs(inst, []string{
		"sudo", "systemctl", "poweroff",
	})...)
	cmd.CombinedOutput() // exit here is expected to be nonzero as the guest dies

	for i := 0; i < timeoutSeconds; i++ {
		if !IsProcessRunning(inst.PID) {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("guest did not power off within %ds", timeoutSeconds)
}

func KillQEMU(pid int) error {
	if !IsProcessRunning(pid) {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if !IsProcessRunning(pid) {
			return nil
		}
		return err
	}
	for i := 0; i < 5; i++ {
		if !IsProcessRunning(pid) {
			return nil
		}
		time.Sleep(time.Second)
	}
	if err := proc.Kill(); err != nil {
		if !IsProcessRunning(pid) {
			return nil
		}
		return err
	}
	return nil
}

func IsProcessRunning(pid int) bool {
	// os.Signal(os.Signal(nil)) is unsupported in this runtime, so a liveness
	// check must not rely on signal 0. Read /proc/<pid>/stat: a running process
	// has state R/S/D/T etc., a zombie has Z. The daemonized qemu is reparented
	// to PID 1, so os.FindProcess may still resolve a zombie before init reaps
	// it; treat Z as not running.
	if pid <= 0 {
		return false
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	fields := strings.Split(string(data), " ")
	if len(fields) < 3 {
		return false
	}
	return fields[2] != "Z"
}
