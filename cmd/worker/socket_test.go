package worker

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type socketClient struct {
	t    *testing.T
	http *http.Client
}

func startSocketServer(t *testing.T, stub *stubLife) (*socketClient, string) {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "w.sock")
	ln, err := Listen(sock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, ln, NewServer(stub, testToken).Handler()) }()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", sock)
	}}
	return &socketClient{t: t, http: &http.Client{Transport: transport}}, sock
}

func (c *socketClient) call(method, path, body, token string) (int, string) {
	c.t.Helper()
	req, err := http.NewRequest(method, "http://worker"+path, strings.NewReader(body))
	if err != nil {
		c.t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(data)
}

func (c *socketClient) expect(want int, method, path, body, token string) string {
	c.t.Helper()
	code, out := c.call(method, path, body, token)
	if code != want {
		c.t.Fatalf("%s %s: got %d, want %d: %s", method, path, code, want, out)
	}
	return out
}

func TestSocketAuthAndMode(t *testing.T) {
	client, sock := startSocketServer(t, newStub())
	info, err := os.Stat(sock)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("socket mode = %v, err %v; want 0600", info.Mode().Perm(), err)
	}
	client.expect(http.StatusUnauthorized, "GET", "/boxes", "", "")
	client.expect(http.StatusUnauthorized, "GET", "/boxes", "", "wrong")
	client.expect(http.StatusOK, "GET", "/boxes", "", testToken)
}

func TestSocketCreateStopSnapshot(t *testing.T) {
	client, _ := startSocketServer(t, newStub())
	body := client.expect(http.StatusCreated, "POST", "/boxes", `{"name":"dev","owner":"ann","cpus":4}`, testToken)
	var box Box
	if err := json.Unmarshal([]byte(body), &box); err != nil || box.SSH.Port != 2223 || box.CPUs != 4 || box.SSH.User != "boite" {
		t.Fatalf("create body: %s (%v)", body, err)
	}
	body = client.expect(http.StatusConflict, "POST", "/boxes/dev/snapshot", `{"tag":"base"}`, testToken)
	if !strings.Contains(body, "not_stopped") {
		t.Fatalf("running snapshot must report not_stopped: %s", body)
	}
	client.expect(http.StatusOK, "POST", "/boxes/dev/stop", "", testToken)
	client.expect(http.StatusCreated, "POST", "/boxes/dev/snapshot", `{"tag":"base"}`, testToken)
	body = client.expect(http.StatusOK, "GET", "/boxes/dev/snapshots", "", testToken)
	if !strings.Contains(body, `"tag":"base"`) {
		t.Fatalf("snapshots: %s", body)
	}
}

func TestListenRefusesRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "w.sock")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(path); err == nil {
		t.Fatal("Listen must not replace a regular file")
	}
}
