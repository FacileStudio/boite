package qemu

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// expiryMarkerSuffix names the expiry-marker entry that pairs with a
// session-scoped materialized key: <KEY>__expires holds unix-epoch seconds.
const expiryMarkerSuffix = "__expires"

// casierTokenState is what an entry boundary concluded about the casier token
// sitting in the guest store.
type casierTokenState int

const (
	casierTokenFresh casierTokenState = iota
	casierTokenExpired
	casierTokenUnmarked
	casierTokenAbsent
)

// expiryMarker returns the guest-store key that carries the expiry marker for
// the materialized key.
func expiryMarker(key string) string {
	return key + expiryMarkerSuffix
}

// parseExpiryMarker reads unix-epoch seconds out of an expiry marker value.
func parseExpiryMarker(marker string) (time.Time, error) {
	sec, err := strconv.ParseInt(strings.TrimSpace(marker), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse expiry marker %q: %w", marker, err)
	}
	return time.Unix(sec, 0), nil
}

// classifyCasierToken maps a probe of the guest store onto a token state. The
// expiry boundary is inclusive: a token whose marker equals the current time
// counts as expired. A missing, empty or unreadable marker fails closed: the
// token counts as unmarked and is dropped rather than silently kept.
func classifyCasierToken(present bool, marker string, now time.Time) casierTokenState {
	if !present {
		return casierTokenAbsent
	}
	exp, err := parseExpiryMarker(marker)
	if err != nil {
		return casierTokenUnmarked
	}
	if !now.Before(exp) {
		return casierTokenExpired
	}
	return casierTokenFresh
}

// casierProbe is what one ssh roundtrip learned about the guest store.
type casierProbe struct {
	KeyPresent bool
	Marker     string
}

// probeCasierToken reports whether the guest store still holds the key and
// the value of its expiry marker, in a single ssh roundtrip. A missing key is
// not an error: it surfaces as present=false with an empty marker.
func probeCasierToken(inst *Instance, key, marker string) (casierProbe, error) {
	presence := "if tiroir get " + shellQuote(key) + " >/dev/null 2>&1; then echo k=yes; else echo k=no; fi"
	value := "printf 'exp=%s\\n' \"$(tiroir get " + shellQuote(marker) + " 2>/dev/null || true)\""
	out, err := SSHOutput(inst, presence+"; "+value)
	if err != nil {
		return casierProbe{}, err
	}
	return parseCasierProbe(out), nil
}

// parseCasierProbe reads the key-presence flag and the marker value out of a
// probe's output.
func parseCasierProbe(out string) casierProbe {
	probe := casierProbe{}
	for line := range strings.SplitSeq(out, "\n") {
		switch {
		case line == "k=yes":
			probe.KeyPresent = true
		case strings.HasPrefix(line, "exp="):
			probe.Marker = strings.TrimPrefix(line, "exp=")
		}
	}
	return probe
}

// expireStaleCasierToken enforces the expiry marker at a boundary: an expired
// or unmarked token is dropped from the guest store (never silently kept) and
// announced once with the reason. It reports whether the store now wants a
// token again; false means a fresh token sits in the store and the boundary
// is done.
func expireStaleCasierToken(inst *Instance, c CasierEnv, warn func(string)) (bool, error) {
	key, marker := c.TokenRef, expiryMarker(c.TokenRef)
	probe, err := probeCasierToken(inst, key, marker)
	if err != nil {
		return false, fmt.Errorf("inspect casier token %s: %w", key, err)
	}
	state := classifyCasierToken(probe.KeyPresent, probe.Marker, time.Now())
	if state == casierTokenFresh {
		return false, nil
	}
	if state == casierTokenAbsent {
		return true, nil
	}
	remote := "tiroir delete " + shellQuote(key) + " && tiroir delete " + shellQuote(marker)
	if err := SSHCommand(inst, []string{remote}); err != nil {
		return false, fmt.Errorf("drop casier token %s: %w", key, err)
	}
	if state == casierTokenExpired {
		warn(fmt.Sprintf("casier token %s expired (ttl %s) and was removed from the sandbox store; it will be re-materialized on the next successful refresh", c.TokenRef, c.TTL))
		return true, nil
	}
	warn(fmt.Sprintf("casier token %s had no readable expiry marker and was removed from the sandbox store; it will be re-materialized on the next successful refresh", c.TokenRef))
	return true, nil
}

// rematerializeCasierToken writes a freshly resolved token plus a new expiry
// marker into the guest store. It reports false when the host cannot supply
// the token (the ref is unset and absent from the host store), leaving the
// key absent for the next boundary to retry.
func rematerializeCasierToken(inst *Instance, key string, ttl time.Duration) (bool, error) {
	token, ok, err := resolveCasierToken(key)
	if err != nil || !ok {
		return ok, err
	}
	expiry := strconv.FormatInt(time.Now().Add(ttl).Unix(), 10)
	remote := "tiroir set " + key + " " + shellQuote(token) + " && tiroir set " + expiryMarker(key) + " " + shellQuote(expiry)
	return true, SSHCommand(inst, []string{remote})
}

// refreshCasierSession runs the session-scoped casier contract at a run/exec
// boundary: drop a stale token, then re-materialize while casier is
// reachable. A pinned token is the user's manual value and is exempt.
func refreshCasierSession(inst *Instance, c CasierEnv, ttl time.Duration, warn func(string)) error {
	if c.TokenRef == "" {
		return fmt.Errorf("env.source casier requires env.casier.token_ref")
	}
	if slices.Contains(inst.PinnedEnv, c.TokenRef) {
		return nil
	}
	needsToken, err := expireStaleCasierToken(inst, c, warn)
	if err != nil || !needsToken {
		return err
	}
	ok, err := rematerializeCasierToken(inst, c.TokenRef, ttl)
	if err != nil {
		return err
	}
	if !ok {
		warn(fmt.Sprintf("casier token %s is not available on the host; the sandbox stays without it until the next successful refresh", c.TokenRef))
	}
	return nil
}
