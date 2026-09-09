package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cloudInitStatusJson runs `cloud-init status --format json` in the guest and
// returns its output verbatim, or "" when the guest is unreachable. The
// structured status carries the per-stage "errors" arrays that a plain
// `cloud-init status` hides, which is what lets create explain an "error"
// state instead of just naming it.
func cloudInitStatusJson(inst *Instance) string {
	cmd := exec.Command("ssh", BuildSSHArgs(inst, []string{"cloud-init status --format json"})...)
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}

// parseCloudInitError pulls the failing module and its message out of the
// structured status output. An error entry is a python repr like
// "('scripts-user', RuntimeError('Runparts: 1 failures (runcmd) in 1 attempted commands'))";
// the module is the first single-quoted token and the message the RuntimeError
// argument. Returns empty strings when nothing user-facing is present.
func parseCloudInitError(raw string) (module, message string) {
	module, message = "", ""
	if i := strings.Index(raw, "('"); i >= 0 {
		rest := raw[i+2:]
		if j := strings.Index(rest, "'"); j >= 0 {
			module = strings.TrimSpace(rest[:j])
		}
	}
	if i := strings.Index(raw, "RuntimeError('"); i >= 0 {
		rest := raw[i+len("RuntimeError('"):]
		if j := strings.Index(rest, "')"); j >= 0 {
			message = strings.TrimSpace(rest[:j])
		}
	}
	return module, message
}

// cloudInitModuleLabel maps a cloud-init module name to something a user can
// act on. scripts-user is where runcmd lives, so that is the headline case;
// package and write_files failures get a plain label rather than an opaque
// module id.
func cloudInitModuleLabel(module string) string {
	if module == "scripts-user" {
		return "runcmd step (cloud-init runcmd)"
	}
	if module == "cc_package_update_apt_install" {
		return "package install (apt)"
	}
	if module == "cc_package_update_apt_upgrade" {
		return "package upgrade (apt)"
	}
	if module == "cc_write_files" {
		return "write_files step"
	}
	if module == "cc_users" {
		return "user creation"
	}
	return module
}

// cloudInitConsoleError scans the QEMU console log (the host-side copy) for the
// last shell error. cloud-init runs runcmd with capture=False, so the failing
// command's stderr goes to the QEMU console and never into cloud-init's own
// log; the console is the only place the concrete failure is recorded. The line
// is stripped of the "[ts] cloud-init[N]: " prefix when present.
func cloudInitConsoleError(inst *Instance) string {
	data, err := os.ReadFile(GetConsoleLogPath(inst.Name))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if !(strings.Contains(line, "not found") ||
			strings.Contains(line, "No such file or directory") ||
			strings.Contains(line, "script returned exit code") ||
			strings.Contains(line, "cannot execute")) {
			continue
		}
		if c := strings.Index(line, ":"); c >= 0 && strings.Contains(line[:c], "]") {
			line = strings.TrimSpace(line[c+1:])
		}
		return line
	}
	return ""
}

// cloudInitFailingToken extracts the bare command name from a console error
// line, e.g. "facile" from "zsh:1: command not found: facile". Used to find the
// offending runcmd in the user's config.
func cloudInitFailingToken(consoleLine string) string {
	for _, prefix := range []string{"command not found: ", "not found: "} {
		if i := strings.Index(consoleLine, prefix); i >= 0 {
			return strings.TrimSpace(consoleLine[i+len(prefix):])
		}
	}
	return ""
}

// cloudInitConfigHint points at the user's config line that triggered the
// failure. A line matches when it mentions the token as a whole shell word,
// preceded by start/quote/space and followed by space/quote/end, so a URL in a
// curl line does not shadow the real command. Falls back to ~/.boite.yml when
// configPath is empty.
func cloudInitConfigHint(configPath, token string) string {
	if token == "" {
		return ""
	}
	p := configPath
	if p == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		p = filepath.Join(home, ".boite.yml")
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return ""
	}
	for i, line := range strings.Split(trimmed, "\n") {
		for _, f := range strings.Split(strings.TrimSpace(line), " ") {
			w := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(f, "'"), "\""))
			if w == token {
				return fmt.Sprintf("%s:%d: %s", p, i+1, strings.TrimSpace(line))
			}
		}
	}
	return ""
}

// cloudInitDiagnostic assembles the explanation of a cloud-init "error" state
// from the guest status, the host console and the user config, as a block of
// indented lines ("" when nothing could be gathered).
func cloudInitDiagnostic(inst *Instance, configPath string) string {
	parts := []string{}
	module, message := parseCloudInitError(cloudInitStatusJson(inst))
	if module != "" {
		label := cloudInitModuleLabel(module)
		if message == "" {
			parts = append(parts, fmt.Sprintf("failure in %s", label))
		} else {
			parts = append(parts, fmt.Sprintf("failure in %s: %s", label, message))
		}
	}
	if console := cloudInitConsoleError(inst); console != "" {
		parts = append(parts, "console: "+console)
		if token := cloudInitFailingToken(console); token != "" {
			if hint := cloudInitConfigHint(configPath, token); hint != "" {
				parts = append(parts, "config: "+hint)
			}
		}
	}
	return strings.Join(parts, "\n  ")
}

// cloudInitErrorSummary frames the "error" state as the warning WaitForCloudInit
// settles on. The sandbox is up and usable, so create does not fail, but the
// warning explains that provisioning stopped short and, when possible, exactly
// where.
func cloudInitErrorSummary(inst *Instance, configPath string, elapsed int) string {
	head := fmt.Sprintf("cloud-init finished with errors (%ds): the sandbox is up but provisioning did not fully complete", elapsed)
	if diag := cloudInitDiagnostic(inst, configPath); diag != "" {
		return head + "\n  " + diag
	}
	return head + ". Run `cloud-init status --format json` inside the VM for details."
}
