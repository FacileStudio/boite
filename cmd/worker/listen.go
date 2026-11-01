package worker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const shutdownGrace = 30 * time.Second

// Listen binds a Unix socket at path with the given mode, replacing a stale
// socket file but refusing to touch any other kind of file. A gid of zero or
// more also hands the socket to that group.
func Listen(path string, mode os.FileMode, gid int) (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create socket dir: %w", err)
	}
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, fmt.Errorf("%s exists and is not a socket", path)
		}
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("remove stale socket: %w", err)
		}
	}
	old := syscall.Umask(0o177)
	ln, err := net.Listen("unix", path)
	syscall.Umask(old)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", path, err)
	}
	if err := secureSocket(path, mode, gid); err != nil {
		ln.Close()
		return nil, err
	}
	return ln, nil
}

func secureSocket(path string, mode os.FileMode, gid int) error {
	if gid >= 0 {
		if err := os.Chown(path, -1, gid); err != nil {
			return fmt.Errorf("chown socket to group %d: %w", gid, err)
		}
	}
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("chmod socket: %w", err)
	}
	return nil
}

// Serve answers requests on ln until ctx ends, then drains for a grace
// period. Boxes are separate processes and keep running after Serve returns.
func Serve(ctx context.Context, ln net.Listener, handler http.Handler) error {
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ln) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-done; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
