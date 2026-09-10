package qemu

import (
	"fmt"
	"maps"
	"sort"

	tiroir "github.com/FacileStudio/tiroir/lib"
)

// materializeEnv resolves the env block into the values to bake into the
// guest store at create. Literals are copied through verbatim; resolve keys
// are read by name from the host tiroir store and skipped when absent, so the
// VM receives only resolved values and never a host secret. The casier source
// is dormant (no token minted or resolution path wired yet) and returns
// nothing.
func materializeEnv(cfg *BoiteConfig) (map[string]string, error) {
	if cfg == nil || cfg.Env == nil {
		return nil, nil
	}
	switch cfg.Env.EffectiveSource() {
	case EnvSourceLocal:
		return materializeLocal(cfg.Env.Local)
	case EnvSourceCasier:
		return nil, nil
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

// RefreshGuestEnv re-materializes the managed env block into the guest store at
// entry (boite run / boite exec), so a VM picks up value changes in ~/.boite.yml
// and the host tiroir store without being recreated. Keys the user pinned via
// 'boite env set' are left alone — a manual value is authoritative over source
// refresh. A failing refresh returns an error so the caller can warn and keep
// the last snapshot, matching the offline-fallback design.
func RefreshGuestEnv(inst *Instance, cfg *BoiteConfig) error {
	if cfg == nil || cfg.Env == nil {
		return nil
	}
	envVars, err := materializeEnv(cfg)
	if err != nil {
		return fmt.Errorf("materialize env: %w", err)
	}
	for _, name := range envRefreshPlan(envVars, inst.PinnedEnv) {
		if err := SSHCommand(inst, []string{"tiroir set " + name + " " + shellQuote(envVars[name])}); err != nil {
			return fmt.Errorf("refresh %s on guest: %w", name, err)
		}
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
