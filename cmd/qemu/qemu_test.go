package qemu

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitForPIDFileWithDelay(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid.txt")

	var writeErr error
	go func() {
		time.Sleep(100 * time.Millisecond)
		data := []byte("12345\n")
		if err := os.WriteFile(pidFile, data, 0644); err != nil {
			writeErr = err
		}
	}()

	start := time.Now()
	pid, err := WaitForPID(pidFile, 10)
	elapsed := time.Since(start)

	if writeErr != nil {
		t.Fatalf("failed to write pid file: %v", writeErr)
	}
	assert.NoError(t, err)
	assert.Equal(t, 12345, pid)
	if elapsed < 100*time.Millisecond {
		t.Fatalf("WaitForPID returned too quickly, expected at least 100ms delay")
	}
}

func TestWaitForPIDFileTimeout(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid.txt")

	_, err := WaitForPID(pidFile, 1)
	assert.Error(t, err)
}

func TestWaitForPIDFileInvalidContent(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid.txt")

	data := []byte("not-a-number\n")
	if err := os.WriteFile(pidFile, data, 0644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	_, err := WaitForPID(pidFile, 5)
	assert.Error(t, err)
}

func TestKillQEMUGracefulTermination(t *testing.T) {
	cmd := exec.Command("sleep", "600")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start child: %v", err)
	}
	pid := cmd.Process.Pid

	if err := KillQEMU(pid); err != nil {
		t.Fatalf("kill qemu: %v", err)
	}

	cmd.Wait()
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		if !IsProcessRunning(pid) {
			return
		}
	}
	t.Fatal("child process still alive after KillQEMU")
}

func TestIsProcessRunningInvalidPid(t *testing.T) {
	if IsProcessRunning(0) {
		t.Fatal("pid 0 must not be running")
	}
	if IsProcessRunning(-1) {
		t.Fatal("negative pid must not be running")
	}
	if IsProcessRunning(999999999) {
		t.Fatal("nonexistent pid must not be running")
	}
}
