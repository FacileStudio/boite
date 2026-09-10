package qemu

import (
	"fmt"
	"maps"

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
