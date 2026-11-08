package qemu

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// PasstDNS is the guest-side nameserver address. passt forwards DNS sent there
// to the host's first resolver from /etc/resolv.conf, so guest DNS works even
// when that resolver listens on host loopback (127.0.0.53), without mapping
// any other guest traffic to the host loopback. It is slirp's DNS address,
// which the baked images hardcode in /etc/resolv.conf. passt is not given
// --dns: that flag stops it reading resolv.conf and the forward target
// becomes 0.0.0.0.
const PasstDNS = "10.0.2.3"

const maxSocketPath = 107

// ErrPasstMissing reports that open mode was requested on a host without the
// passt binary.
var ErrPasstMissing = errors.New("passt not found in PATH: vm.net open needs passt to keep the guest " +
	"off the host loopback (install the passt package), or set vm.net: open-slirp to accept the unsafe slirp " +
	"network where the guest reaches host 127.0.0.1 services through 10.0.2.2")

// PasstConfig locates the per-VM passt process: the Unix socket QEMU attaches
// to, the PID file passt writes, and the log that keeps its output.
type PasstConfig struct {
	Socket  string
	PIDFile string
	Log     string
}

// NewPasstConfig returns the passt socket, PID file and log paths kept in an
// instance directory.
func NewPasstConfig(instanceDir string) *PasstConfig {
	return &PasstConfig{
		Socket:  filepath.Join(instanceDir, "passt.sock"),
		PIDFile: filepath.Join(instanceDir, "passt.pid"),
		Log:     filepath.Join(instanceDir, "passt.log"),
	}
}

// PasstArgs builds the passt command line for one VM. --no-map-gw keeps the
// gateway from aliasing the host loopback, --one-off makes passt exit once
// QEMU disconnects, and the only inbound forward is SSH on the bind address.
func PasstArgs(p *PasstConfig, sshPort int) []string {
	return []string{
		"--one-off",
		"--no-map-gw",
		"--socket", p.Socket,
		"--pid", p.PIDFile,
		"--tcp-ports", fmt.Sprintf("%s/%d:22", SSHBindAddr(), sshPort),
		"--dns-forward", PasstDNS,
	}
}

// netdevArgs returns the QEMU NIC flags: a stream netdev on the passt socket
// when cfg.Passt is set, otherwise slirp with the SSH hostfwd, restricted when
// cfg.RestrictNet is set.
func netdevArgs(cfg QEMUConfig) []string {
	restrict := "off"
	if cfg.RestrictNet {
		restrict = "on"
	}
	netdev := fmt.Sprintf("user,id=net0,restrict=%s,hostfwd=tcp:%s:%d-:22", restrict, SSHBindAddr(), cfg.HostFwdPort)
	if cfg.Passt != nil {
		netdev = "stream,id=net0,server=off,addr.type=unix,addr.path=" + cfg.Passt.Socket
	}
	return []string{"-netdev", netdev, "-device", "virtio-net-pci,netdev=net0"}
}

// StartPasst launches passt for p, forwarding sshPort to the guest, and waits
// until its socket and PID file exist. A nil p is a no-op. A stale passt from
// an earlier boot of the same VM is stopped first.
func StartPasst(p *PasstConfig, sshPort int) error {
	if p == nil {
		return nil
	}
	bin, err := exec.LookPath("passt")
	if err != nil {
		return ErrPasstMissing
	}
	if len(p.Socket) > maxSocketPath {
		return fmt.Errorf("passt socket path %s is longer than %d bytes: use a shorter BOITE_DIR", p.Socket, maxSocketPath)
	}
	if err := StopPasst(p.PIDFile); err != nil {
		return err
	}
	os.Remove(p.Socket)
	logFile, err := os.Create(p.Log)
	if err != nil {
		return fmt.Errorf("create passt log: %w", err)
	}
	defer logFile.Close()
	cmd := exec.Command(bin, PasstArgs(p, sshPort)...)
	cmd.Stdout, cmd.Stderr = logFile, logFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("start passt: %w (see %s)", err, p.Log)
	}
	return waitForPasst(p, 50)
}

func waitForPasst(p *PasstConfig, attempts int) error {
	for range attempts {
		if _, err := os.Stat(p.Socket); err == nil && passtPID(p.PIDFile) > 0 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	StopPasst(p.PIDFile)
	return fmt.Errorf("passt socket %s did not appear (see %s)", p.Socket, p.Log)
}

// StopPasst terminates the passt whose PID is recorded in pidFile and removes
// the file. It only signals a live process named passt, so a recycled PID is
// never killed. A missing PID file means there is nothing to stop.
func StopPasst(pidFile string) error {
	pid := passtPID(pidFile)
	if pid > 0 {
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil && IsProcessRunning(pid) {
			return fmt.Errorf("stop passt %d: %w", pid, err)
		}
		if err := WaitForProcessExit(pid, 5); err != nil {
			syscall.Kill(pid, syscall.SIGKILL)
		}
	}
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove passt pid file: %w", err)
	}
	return nil
}

func passtPID(pidFile string) int {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || !IsProcessRunning(pid) {
		return 0
	}
	comm, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil || !strings.HasPrefix(strings.TrimSpace(string(comm)), "passt") {
		return 0
	}
	return pid
}
