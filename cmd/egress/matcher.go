package egress

import "fmt"

// ValidateHosts parses every entry, drops duplicates, and returns the
// normalized strings. It fails on the first bad entry or when the list is
// longer than MaxExtraHosts.
func ValidateHosts(entries []string) ([]string, error) {
	if len(entries) > MaxExtraHosts {
		return nil, fmt.Errorf("at most %d extra hosts", MaxExtraHosts)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		r, err := ParseRule(e)
		if err != nil {
			return nil, err
		}
		if s := r.String(); !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out, nil
}

// NewMatcher builds a Matcher from the built-in defaults plus extra entries.
// Entries that do not parse are an error.
func NewMatcher(extra []string) (*Matcher, error) {
	m := &Matcher{}
	for _, e := range append(append([]string{}, DefaultHosts()...), extra...) {
		r, err := ParseRule(e)
		if err != nil {
			return nil, err
		}
		m.rules = append(m.rules, r)
	}
	return m, nil
}

// Allowed reports whether host and port are on the list. IP literals and
// names that fail the host syntax checks are always refused.
func (m *Matcher) Allowed(host string, port int) bool {
	h, err := normalizeHost(host)
	if err != nil {
		return false
	}
	for _, r := range m.rules {
		if r.allows(h, port) {
			return true
		}
	}
	return false
}
