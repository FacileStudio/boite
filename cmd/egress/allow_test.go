package egress

import (
	"fmt"
	"testing"
)

func TestMatcher(t *testing.T) {
	m, err := NewMatcher([]string{"api.example.com", "*.cdn.example.org", "git.example.net:9418", "Mixed.Example.COM."})
	if err != nil {
		t.Fatal(err)
	}
	cases := matcherCases()
	for _, c := range cases {
		if got := m.Allowed(c.host, c.port); got != c.want {
			t.Errorf("Allowed(%q, %d) = %v, want %v", c.host, c.port, got, c.want)
		}
	}
}

func TestParseRuleRejects(t *testing.T) {
	bad := []string{
		"", "localhost", "*.com", "*", "10.0.0.1", "[::1]", "::1", "a.com:0", "a.com:99999", "a.com:x",
		"a.com/path", "http://a.com", "a b.com", "-a.com", "a-.com", "a..com", "*.*.com", "a.123", "a.0x1f",
		"user@a.com", "a.com:80:81",
	}
	for _, e := range bad {
		if _, err := ParseRule(e); err == nil {
			t.Errorf("ParseRule(%q) accepted", e)
		}
	}
}

func TestValidateHosts(t *testing.T) {
	got, err := ValidateHosts([]string{"A.com", "a.com", "*.B.com:8080"})
	if err != nil || len(got) != 2 || got[0] != "a.com" || got[1] != "*.b.com:8080" {
		t.Fatalf("ValidateHosts = %v, %v", got, err)
	}
	many := make([]string, MaxExtraHosts+1)
	for i := range many {
		many[i] = fmt.Sprintf("h%d.example.com", i)
	}
	if _, err := ValidateHosts(many); err == nil {
		t.Fatal("too many hosts accepted")
	}
}

func TestDefaultHostsParse(t *testing.T) {
	if _, err := NewMatcher(nil); err != nil {
		t.Fatal(err)
	}
}

func TestMaybeServeIgnoresOtherArgs(t *testing.T) {
	for _, args := range [][]string{{"boite"}, {"boite", "list"}, {"boite", HelperArg}, {"boite", HelperArg, "a"}} {
		if served, _ := MaybeServe(args); served {
			t.Fatalf("MaybeServe(%v) served", args)
		}
	}
	if served, code := MaybeServe([]string{"boite", HelperArg, "/nonexistent/hosts", "/dev/null"}); !served || code != 1 {
		t.Fatalf("missing hosts file: served=%v code=%d", served, code)
	}
}

type matcherCase struct {
	host string
	port int
	want bool
}

func matcherCases() []matcherCase {
	return append(hostCases(), suffixCases()...)
}

func hostCases() []matcherCase {
	return []matcherCase{
		{"github.com", 443, true},
		{"GitHub.com", 443, true},
		{"github.com.", 443, true},
		{"github.com", 80, true},
		{"github.com", 22, false},
		{"github.com", 8443, false},
	}
}

func suffixCases() []matcherCase {
	return []matcherCase{
		{"evilgithub.com", 443, false},
		{"github.com.evil.com", 443, false},
		{"notgithub.com", 443, false},
		{"sub.github.com", 443, false},
		{"registry.npmjs.org", 443, true},
		{"pypi.org", 443, true},
		{"files.pythonhosted.org", 443, true},
		{"api.example.com", 443, true},
		{"x.api.example.com", 443, false},
		{"mixed.example.com", 443, true},
		{"cdn.example.org", 443, false},
		{"a.cdn.example.org", 443, true},
		{"a.b.cdn.example.org", 443, true},
		{"evilcdn.example.org", 443, false},
		{"git.example.net", 9418, true},
		{"git.example.net", 443, false},
		{"1.2.3.4", 443, false},
		{"127.0.0.1", 80, false},
		{"0x7f.0.0.1", 80, false},
		{"2130706433", 80, false},
		{"::1", 443, false},
		{"[::1]", 443, false},
		{"localhost", 80, false},
		{"", 443, false},
		{"github.com\x00.evil.com", 443, false},
		{"git hub.com", 443, false},
	}
}
