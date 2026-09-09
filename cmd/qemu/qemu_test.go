package qemu

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitForPIDFileWithDelay(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid.txt")

	go func() {
		time.Sleep(100 * time.Millisecond)
		data := []byte("12345\n")
		if err := os.WriteFile(pidFile, data, 0644); err != nil {
			t.Fatalf("failed to write pid file: %v", err)
		}
	}()

	start := time.Now()
	pid, err := WaitForPID(pidFile, 10)
	elapsed := time.Since(start)

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
