// Package egress decides which hosts an allowlist box may reach and serves
// the host-side HTTP proxy that enforces it.
package egress

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

// MaxExtraHosts caps the per-box extra entries a caller may pass.
const MaxExtraHosts = 64

const maxHostLen = 253

// Rule is one allowed destination: an exact host or a "*." subdomain
// wildcard, optionally pinned to one port. Without a port it covers 80 and
// 443.
type Rule struct {
	Host     string
	Wildcard bool
	Port     int
}

// Matcher answers whether a host and port are on the allow-list.
type Matcher struct {
	rules []Rule
}

// ParseRule parses "host", "*.host" or "host:port" into a Rule. It lowercases,
// drops one trailing dot, and refuses IP literals, single-label names, bare
// wildcards such as "*.com", and anything that is not letters, digits and
// dashes.
func ParseRule(entry string) (Rule, error) {
	s := strings.ToLower(strings.TrimSpace(entry))
	var r Rule
	if strings.HasPrefix(s, "[") || strings.Count(s, ":") > 1 {
		return r, fmt.Errorf("%q: IP literals are not allowed", entry)
	}
	if host, port, ok := strings.Cut(s, ":"); ok {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return r, fmt.Errorf("%q: invalid port", entry)
		}
		s, r.Port = host, n
	}
	if rest, ok := strings.CutPrefix(s, "*."); ok {
		s, r.Wildcard = rest, true
	}
	host, err := normalizeHost(s)
	if err != nil {
		return r, fmt.Errorf("%q: %w", entry, err)
	}
	r.Host = host
	return r, nil
}

func normalizeHost(h string) (string, error) {
	h = strings.TrimSuffix(strings.ToLower(h), ".")
	if h == "" || len(h) > maxHostLen {
		return "", fmt.Errorf("invalid host length")
	}
	if _, err := netip.ParseAddr(h); err == nil {
		return "", fmt.Errorf("IP literals are not allowed")
	}
	labels := strings.Split(h, ".")
	if len(labels) < 2 {
		return "", fmt.Errorf("host needs at least two labels")
	}
	for _, l := range labels {
		if !validLabel(l) {
			return "", fmt.Errorf("invalid label %q", l)
		}
	}
	if tld := labels[len(labels)-1]; !strings.ContainsAny(tld, "abcdefghijklmnopqrstuvwxyz") || isHexNumber(tld) {
		return "", fmt.Errorf("top-level label %q looks like a number", tld)
	}
	return h, nil
}

func validLabel(l string) bool {
	if l == "" || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' {
		return false
	}
	for _, c := range l {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}

func isHexNumber(s string) bool {
	rest, ok := strings.CutPrefix(s, "0x")
	if !ok {
		return false
	}
	_, err := strconv.ParseUint("0"+rest, 16, 64)
	return err == nil
}

// String renders the rule the way ParseRule reads it.
func (r Rule) String() string {
	s := r.Host
	if r.Wildcard {
		s = "*." + s
	}
	if r.Port != 0 {
		s += ":" + strconv.Itoa(r.Port)
	}
	return s
}

func (r Rule) allows(host string, port int) bool {
	if r.Port == 0 && port != 80 && port != 443 || r.Port != 0 && r.Port != port {
		return false
	}
	if r.Wildcard {
		return strings.HasSuffix(host, "."+r.Host)
	}
	return host == r.Host
}
