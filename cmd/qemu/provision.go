package qemu

import (
	"fmt"
	"os/exec"
	"strings"
)

// Provision applies the per-instance provision block right after firstboot:
// apt packages via the boite user's passwordless sudo, then each command as a
// shell line. An absent or empty block is a no-op. Output is captured, not
// streamed, so apt progress cannot corrupt the single progress line; a failing
// step returns an error that includes the captured output.
func Provision(inst *Instance, cfg *BoiteConfig) error {
	if cfg == nil || cfg.Provision == nil || cfg.Provision.Empty() {
		return nil
	}

	ProgressPhase("Provisioning instance")

	hasPackages := len(cfg.Provision.Packages) > 0
	if err := provisionPackages(inst, cfg.Provision.Packages); err != nil {
		return err
	}

	for i, command := range cfg.Provision.Commands {
		frac := commandFrac(hasPackages, i, len(cfg.Provision.Commands))
		ProgressTick(fmt.Sprintf("running command %d/%d", i+1, len(cfg.Provision.Commands)), frac)
		if err := runProvisionStep(inst, []string{"sh", "-lc", shellQuote(command)}); err != nil {
			return fmt.Errorf("provision command %d: %w", i+1, err)
		}
	}

	ProgressDone("provisioning complete")
	return nil
}

// shellQuote wraps a command so it survives the trip through OpenSSH and the
// remote login shell as a single word. ssh rejoins its command arguments with
// spaces into one line, which the remote shell re-parses; an unquoted command
// therefore has its embedded spaces, pipes and multiline string torn apart. A
// single-quoted whole command is kept intact as sh -c's argument.
func shellQuote(cmd string) string {
	return "'" + strings.ReplaceAll(cmd, "'", `'\''`) + "'"
}

// provisionPackages refreshes apt and installs the requested packages in one
// invocation.
func provisionPackages(inst *Instance, packages []string) error {
	if len(packages) == 0 {
		return nil
	}
	ProgressTick("apt: refreshing package lists", 0)
	if err := runProvisionStep(inst, []string{"sudo", "apt-get", "update"}); err != nil {
		return fmt.Errorf("apt-get update: %w", err)
	}
	ProgressTick("apt: installing custom packages", 0.4)
	install := append([]string{"sudo", "apt-get", "install", "-y", "--no-install-recommends"}, packages...)
	if err := runProvisionStep(inst, install); err != nil {
		return fmt.Errorf("apt-get install %s: %w", strings.Join(packages, " "), err)
	}
	return nil
}

// commandFrac maps a command's position onto the determinate bar, skipping the
// first half when a package step already used it.
func commandFrac(hasPackages bool, i, total int) float64 {
	base := 0.0
	if hasPackages {
		base = 0.5
	}
	return base + (1-base)*float64(i)/float64(total)
}

// runProvisionStep runs one remote command and returns its combined output on
// failure so the caller can explain what went wrong.
func runProvisionStep(inst *Instance, args []string) error {
	cmd := exec.Command("ssh", BuildSSHArgs(inst, args)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
