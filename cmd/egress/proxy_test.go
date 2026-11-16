package egress

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
)

func testProxy(t *testing.T, upstream string) (*httptest.Server, *[]string) {
	t.Helper()
	m, err := NewMatcher([]string{"ok.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	var logs []string
	p := NewProxy(m, func(f string, a ...any) { logs = append(logs, fmt.Sprintf(f, a...)) })
	p.dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, upstream)
	}
	srv := httptest.NewServer(p)
	t.Cleanup(srv.Close)
	return srv, &logs
}

func connect(t *testing.T, proxy, target string) (*bufio.Reader, net.Conn, int) {
	t.Helper()
	c, err := net.Dial("tcp", strings.TrimPrefix(proxy, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	fmt.Fprintf(c, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target)
	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	return br, c, resp.StatusCode
}

func TestProxyConnectTunnels(t *testing.T) {
	up, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer up.Close()
	go func() {
		c, err := up.Accept()
		if err == nil {
			io.Copy(c, c)
			c.Close()
		}
	}()
	srv, logs := testProxy(t, up.Addr().String())
	br, c, code := connect(t, srv.URL, "ok.example.com:443")
	if code != 200 {
		t.Fatalf("allowed CONNECT = %d", code)
	}
	fmt.Fprint(c, "ping")
	buf := make([]byte, 4)
	if _, err := io.ReadFull(br, buf); err != nil || string(buf) != "ping" {
		t.Fatalf("echo = %q, %v", buf, err)
	}
	for _, target := range []string{"evil.com:443", "ok.example.com:22", "127.0.0.1:443", "10.0.0.1:80"} {
		if _, _, code := connect(t, srv.URL, target); code != 403 {
			t.Errorf("CONNECT %s = %d, want 403", target, code)
		}
	}
	if len(*logs) != 4 {
		t.Fatalf("denials logged = %v", *logs)
	}
}

func TestProxyForwardsPlainHTTP(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello %s %q", r.URL.Path, r.Header.Get("X-Forwarded-For"))
	}))
	defer up.Close()
	srv, _ := testProxy(t, strings.TrimPrefix(up.URL, "http://"))
	get := func(target string) (int, string) {
		req, _ := http.NewRequest("GET", target, nil)
		pu := srv.URL
		tr := &http.Transport{Proxy: func(*http.Request) (*url.URL, error) { return url.Parse(pu) }}
		resp, err := tr.RoundTrip(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	if code, body := get("http://ok.example.com/x"); code != 200 || body != `hello /x ""` {
		t.Fatalf("allowed GET = %d %q", code, body)
	}
	if code, _ := get("http://evil.com/x"); code != 403 {
		t.Fatalf("denied GET = %d", code)
	}
	if code, _ := get("http://ok.example.com:8080/x"); code != 403 {
		t.Fatalf("denied port GET = %d", code)
	}
}

func TestPublicAddr(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8": true, "140.82.112.3": true, "2606:4700::1": true,
		"127.0.0.1": false, "10.1.2.3": false, "192.168.1.1": false, "172.16.0.1": false,
		"169.254.169.254": false, "100.64.0.1": false, "0.0.0.0": false, "::1": false,
		"fe80::1": false, "fd00::1": false, "::ffff:127.0.0.1": false, "224.0.0.1": false,
	}
	for s, want := range cases {
		if got := publicAddr(netip.MustParseAddr(s)); got != want {
			t.Errorf("publicAddr(%s) = %v, want %v", s, got, want)
		}
	}
	if err := publicOnly("tcp", "127.0.0.1:80", nil); err == nil {
		t.Error("publicOnly accepted loopback")
	}
	if err := publicOnly("tcp", "[::1]:80", nil); err == nil {
		t.Error("publicOnly accepted ::1")
	}
}
