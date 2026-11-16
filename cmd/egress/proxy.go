package egress

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"strconv"
	"syscall"
	"time"
)

const dialTimeout = 15 * time.Second

// Proxy is an HTTP forward proxy that serves CONNECT and plain http:// requests
// only for hosts the Matcher allows. It dials the resolved address itself and
// refuses any address that is not a public unicast one, so an allowed name
// cannot be pointed at the host or its LAN.
type Proxy struct {
	Matcher *Matcher
	Logf    func(format string, args ...any)
	dial    func(ctx context.Context, network, addr string) (net.Conn, error)
	reverse *httputil.ReverseProxy
}

// NewProxy returns a Proxy enforcing m. logf may be nil.
func NewProxy(m *Matcher, logf func(string, ...any)) *Proxy {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	d := &net.Dialer{Timeout: dialTimeout, Control: publicOnly}
	p := &Proxy{Matcher: m, Logf: logf, dial: d.DialContext}
	p.reverse = &httputil.ReverseProxy{
		Director:  func(r *http.Request) { r.Header["X-Forwarded-For"] = nil },
		Transport: &http.Transport{DialContext: func(ctx context.Context, n, a string) (net.Conn, error) { return p.dial(ctx, n, a) }, ResponseHeaderTimeout: 60 * time.Second},
	}
	return p
}

func publicOnly(_, address string, _ syscall.RawConn) error {
	ap, err := netip.ParseAddrPort(address)
	if err != nil || !publicAddr(ap.Addr()) {
		return fmt.Errorf("address %s is not a public unicast address", address)
	}
	return nil
}

func publicAddr(a netip.Addr) bool {
	a = a.Unmap()
	cgnat := netip.MustParsePrefix("100.64.0.0/10")
	return a.IsGlobalUnicast() && !a.IsPrivate() && !cgnat.Contains(a)
}

// ServeHTTP implements http.Handler.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.connect(w, r)
		return
	}
	p.forward(w, r)
}

func (p *Proxy) check(w http.ResponseWriter, hostport, defPort string) (string, bool) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		host, port = hostport, defPort
	}
	n, perr := strconv.Atoi(port)
	if perr != nil || !p.Matcher.Allowed(host, n) {
		p.Logf("deny %s", hostport)
		http.Error(w, "egress: host not on the allow-list", http.StatusForbidden)
		return "", false
	}
	return net.JoinHostPort(host, port), true
}

func (p *Proxy) forward(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() || r.URL.Scheme != "http" {
		http.Error(w, "egress: proxy requests need an absolute http URL", http.StatusBadRequest)
		return
	}
	if _, ok := p.check(w, r.URL.Host, "80"); !ok {
		return
	}
	r.Header.Del("Proxy-Connection")
	p.reverse.ServeHTTP(w, r)
}

func (p *Proxy) connect(w http.ResponseWriter, r *http.Request) {
	target, ok := p.check(w, r.Host, "443")
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), dialTimeout)
	defer cancel()
	up, err := p.dial(ctx, "tcp", target)
	if err != nil {
		p.Logf("dial %s: %v", target, err)
		http.Error(w, "egress: upstream unreachable", http.StatusBadGateway)
		return
	}
	defer up.Close()
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "egress: cannot hijack", http.StatusInternalServerError)
		return
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		return
	}
	defer conn.Close()
	io.WriteString(conn, "HTTP/1.1 200 Connection Established\r\n\r\n")
	pipe(conn, rw, up)
}
