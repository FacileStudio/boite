package cmd

import (
	"slices"

	"github.com/FacileStudio/boite/cmd/qemu"
)

// pinEnv marks a key as manually set so source refresh leaves it alone.
func pinEnv(inst *qemu.Instance, name string) error {
	if slices.Contains(inst.PinnedEnv, name) {
		return nil
	}
	inst.PinnedEnv = append(inst.PinnedEnv, name)
	return qemu.SaveInstanceState(inst)
}

// unpinEnv forgets a key after it is deleted, so a managed value can be restored
// by the next entry refresh.
func unpinEnv(inst *qemu.Instance, name string) error {
	rest := inst.PinnedEnv[:0]
	changed := false
	for _, p := range inst.PinnedEnv {
		if p != name {
			rest = append(rest, p)
		} else {
			changed = true
		}
	}
	if !changed {
		return nil
	}
	inst.PinnedEnv = rest
	return qemu.SaveInstanceState(inst)
}

// RefreshManagedEnv re-materializes the env block of configPath into the named
// instance's guest store. It loads the instance and config fresh from disk so
// callers (run/exec) need no prior state. Non-fatal by contract: callers warn
// on error and keep the last guest snapshot.
func RefreshManagedEnv(name, configPath string) error {
	inst, err := qemu.LoadInstanceState(name)
	if err != nil {
		return err
	}
	cfg, err := qemu.LoadBoiteConfig(configPath)
	if err != nil {
		return err
	}
	return qemu.RefreshGuestEnv(inst, cfg)
}
