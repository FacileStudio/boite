package egress

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// HelperArg is the first argument that turns a boite binary into the
// per-connection proxy QEMU spawns for a guestfwd rule.
const HelperArg = "__egress-proxy"

// MaybeServe runs the stdio proxy when args select it and reports whether it
// did. Both binaries call it first thing in main, then exit.
func MaybeServe(args []string) (bool, int) {
	if len(args) != 4 || args[1] != HelperArg {
		return false, 0
	}
	if err := ServeStdio(args[2], args[3]); err != nil {
		fmt.Fprintln(os.Stderr, "egress:", err)
		return true, 1
	}
	return true, 0
}

// LoadHosts reads the extra hosts file: one entry per line, blank lines
// ignored.
func LoadHosts(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for l := range strings.SplitSeq(string(data), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out, nil
}

// WriteHosts stores the extra hosts, one per line.
func WriteHosts(path string, hosts []string) error {
	return os.WriteFile(path, []byte(strings.Join(hosts, "\n")+"\n"), 0o600)
}

// ServeStdio serves one guest connection: QEMU hands it to this process as
// stdin and stdout. Denied requests are appended to logPath.
func ServeStdio(hostsPath, logPath string) error {
	extra, err := LoadHosts(hostsPath)
	if err != nil {
		return err
	}
	m, err := NewMatcher(extra)
	if err != nil {
		return err
	}
	conn := stdioConn()
	srv := &http.Server{Handler: NewProxy(m, fileLogger(logPath)), ReadHeaderTimeout: 30 * time.Second}
	err = srv.Serve(newOneShot(conn))
	if errors.Is(err, errServed) {
		return nil
	}
	return err
}

func stdioConn() net.Conn {
	if c, err := net.FileConn(os.Stdin); err == nil {
		return c
	}
	return &pipeConn{File: os.Stdin, w: os.Stdout}
}

func fileLogger(path string) func(string, ...any) {
	return func(format string, args ...any) {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return
		}
		defer f.Close()
		fmt.Fprintf(f, "%s %s\n", time.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, args...))
	}
}
