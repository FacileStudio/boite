package qemu

import (
	"fmt"
	"maps"
	"os"
	"sort"
	"strconv"
	"time"

	tiroir "github.com/FacileStudio/tiroir/lib"
)

// materializeEnv resolves the env block into the values to bake into the
// guest store at create. Literals are copied through verbatim; resolve keys
// are read by name from the host tiroir store and skipped when absent, so the
// VM receives only resolved values and never a host secret. The casier source
// materializes the scoped token under token_ref, with an expiry marker when
// env.casier.ttl is set.
func materializeEnv(cfg *BoiteConfig) (map[string]string, error) {
	if cfg == nil || cfg.Env == nil {
		return nil, nil
	}
	switch cfg.Env.EffectiveSource() {
	case EnvSourceLocal:
		return materializeLocal(cfg.Env.Local)
	case EnvSourceCasier:
		return materializeCasier(cfg.Env.Casier)
	default:
		return nil, fmt.Errorf("unknown env source %q", cfg.Env.Source)
	}
}

// materializeLocal copies the literal vars and resolves each name in Resolve
// from the host tiroir store.
func materializeLocal(local LocalEnv) (map[string]string, error) {
	out := maps.Clone(local.Vars)
	if out == nil {
		out = map[string]string{}
	}
	if len(local.Resolve) == 0 {
		return out, nil
	}
	host, err := tiroir.New()
	if err != nil {
		return nil, fmt.Errorf("open host tiroir store: %w", err)
	}
	for _, name := range local.Resolve {
		v, ok, err := host.Get(name)
		if err != nil {
			return nil, fmt.Errorf("resolve %s from host store: %w", name, err)
		}
		if ok {
			out[name] = v
		}
	}
	return out, nil
}

// resolveCasierToken reads the scoped casier token named by token_ref: the
// host environment variable first, then the host tiroir store.
func resolveCasierToken(ref string) (string, bool, error) {
	if v := os.Getenv(ref); v != "" {
		return v, true, nil
	}
	host, err := tiroir.New()
	if err != nil {
		return "", false, fmt.Errorf("open host tiroir store: %w", err)
	}
	v, ok, err := host.Get(ref)
	if err != nil {
		return "", false, fmt.Errorf("resolve %s from host store: %w", ref, err)
	}
	return v, ok, nil
}

// materializeCasier resolves the casier source into guest-store entries: the
// scoped token under token_ref, plus an expiry marker when ttl is set.
func materializeCasier(c CasierEnv) (map[string]string, error) {
	if c.TokenRef == "" {
		return nil, fmt.Errorf("env.source casier requires env.casier.token_ref")
	}
	token, ok, err := resolveCasierToken(c.TokenRef)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("casier token %s is not available on the host (set $%s or store %s in the host tiroir store)", c.TokenRef, c.TokenRef, c.TokenRef)
	}
	out := map[string]string{c.TokenRef: token}
	ttl, err := c.tokenTTL()
	if err != nil {
		return nil, err
	}
	if ttl > 0 {
		out[expiryMarker(c.TokenRef)] = strconv.FormatInt(time.Now().Add(ttl).Unix(), 10)
	}
	return out, nil
}

// RefreshGuestEnv re-materializes the managed env block into the guest store at
// entry (boite run / boite exec), so a VM picks up value changes in ~/.boite.yml
// and the host tiroir store without being recreated. Keys the user pinned via
// 'boite env set' are left alone — a manual value is authoritative over source
// refresh. A failing refresh returns an error so the caller can warn and keep
// the last snapshot, matching the offline-fallback design. A casier source
// with env.casier.ttl set instead enforces session-scoped tokens: the expiry
// marker is checked at every boundary and warn receives the notices (nil warn
// keeps them silent).
func RefreshGuestEnv(inst *Instance, cfg *BoiteConfig, warn func(string)) error {
	if cfg == nil || cfg.Env == nil {
		return nil
	}
	if warn == nil {
		warn = func(string) {}
	}
	if cfg.Env.EffectiveSource() == EnvSourceCasier {
		ttl, err := cfg.Env.Casier.tokenTTL()
		if err != nil {
			return err
		}
		if ttl > 0 {
			return refreshCasierSession(inst, cfg.Env.Casier, ttl, warn)
		}
	}
	return refreshManagedKeys(inst, cfg)
}

// refreshManagedKeys re-materializes every managed key into the guest store in
// a single ssh roundtrip, skipping keys pinned with a manual 'boite env set'.
func refreshManagedKeys(inst *Instance, cfg *BoiteConfig) error {
	envVars, err := materializeEnv(cfg)
	if err != nil {
		return fmt.Errorf("materialize env: %w", err)
	}
	plan := envRefreshPlan(envVars, inst.PinnedEnv)
	if len(plan) == 0 {
		return nil
	}
	var remote string
	for i, name := range plan {
		set := "tiroir set " + name + " " + shellQuote(envVars[name])
		if i == 0 {
			remote = set
		} else {
			remote += " && " + set
		}
	}
	if err := SSHCommand(inst, []string{remote}); err != nil {
		return fmt.Errorf("refresh env on guest: %w", err)
	}
	return nil
}

// envRefreshPlan returns the managed keys to write on entry: every source key
// that the instance has not pinned with a manual 'boite env set'.
func envRefreshPlan(envVars map[string]string, pinned []string) []string {
	skip := make(map[string]bool, len(pinned))
	for _, p := range pinned {
		skip[p] = true
	}
	var plan []string
	for name := range envVars {
		if !skip[name] {
			plan = append(plan, name)
		}
	}
	sort.Strings(plan)
	return plan
}
