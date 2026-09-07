package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func detectAcceleration() (string, []string, string) {
	switch runtime.GOOS {
	case "linux":
		if f, err := os.Open("/dev/kvm"); err == nil {
			f.Close()
			return "-enable-kvm", nil, "host"
		}
		return "", []string{"-machine", "accel=tcg"}, "qemu64"
	case "darwin":
		return "", []string{"-machine", "accel=hvf"}, "qemu64"
	default:
		return "", []string{"-machine", "accel=tcg"}, "qemu64"
	}
}

func defaultVMConfig() *VMConfig {
	return &VMConfig{
		CPUs:   2,
		Memory: "2G",
	}
}

func BuildQEMUArgs(baseImage, overlayPath, seedISOPath string, sshBasePort int, pidFile string, vmCfg *VMConfig) []string {
	accelFlag, accelExtra, cpuModel := detectAcceleration()
	sshListenPort := sshBasePort + 1
	args := []string{}

	if vmCfg == nil {
		vmCfg = defaultVMConfig()
	}

	if accelFlag != "" {
		args = append(args, accelFlag)
	}
	args = append(args, accelExtra...)
	netdevFmt := "user,id=net0,net=192.168.42.0/24,dhcpstart=192.168.42.10,restrict=off,hostfwd=tcp:127.0.0.1:%d-:22"
	args = append(args,
		"-cpu", cpuModel,
		"-m", vmCfg.Memory,
		"-smp", fmt.Sprintf("%d", vmCfg.CPUs),
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", overlayPath),
		"-drive", fmt.Sprintf("file=%s,format=raw,if=virtio,readonly=on", seedISOPath),
		"-netdev", fmt.Sprintf(netdevFmt, sshListenPort),
		"-device", "virtio-net-pci,netdev=net0",
		"-display", "none",
		"-daemonize",
		"-pidfile", pidFile,
		"-name", "boite-vm",
	)
	return args
}

func StartQEMU(baseImage, overlayPath, seedISOPath string, sshPort int, pidFile string, vmCfg *VMConfig) (*exec.Cmd, error) {
	args := BuildQEMUArgs(baseImage, overlayPath, seedISOPath, sshPort, pidFile, vmCfg)
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
	}
	return 0, fmt.Errorf("timeout waiting for PID file")
}

func KillQEMU(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func IsProcessRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(os.Signal(nil))
	return err == nil
}
