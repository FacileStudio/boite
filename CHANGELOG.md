# Changelog

All notable changes to this project are documented here. The format is
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html). While on
`0.x`, a breaking change bumps the minor.

## [Unreleased]

## [0.1.14] — 2026-09-09

### Changed

- Reworked provisioning output into a small animated TTY region on `create` (and `start`, `stop`, `rm`): each provisioning step shows a spinner, the current wait (QEMU boot, SSH, cloud-init) runs a live progress bar with a percentage, and the settled steps scroll above it. When stderr is piped (not a TTY), output degrades to plain phase/done lines with no animation. A failed wait settles as a red `✗` line.

### Fixed

- `create` no longer finishes before cloud-init has completed. A cloud-init timeout used to be downgraded to a "may not have completed" warning and the sandbox was declared created anyway, so the user could enter a VM that was still bootstrapping (a mostly-configured shell could drop with exit status 127). Now `create` blocks until cloud-init reports `status: done` and fails with a clear error on timeout. While it waits, the live tick surfaces cloud-init's own status plus the latest VM console line, so you can see what it is doing.
- Cloud-init waiting no longer fails a healthy first boot. `create` used to abort after a fixed 180s wall clock even while cloud-init was legitimately still `running` (a fresh VM's package install and toolchain bootstrap can take several minutes), reporting "cloud-init never finished" for a box that actually succeeded moments later. It now fails only on a terminal cloud-init state (`error`, `disabled`, etc.) or when the guest never confirms cloud-init started; a confirmed running boot keeps waiting until `status: done`.
- `boite run` no longer reports a session drop as a connection error. ssh signals a failed connection with exit 255; the remote shell's own exit status (0 on a clean exit, or a nonzero status such as 127) means the session ran, so only a genuine connection failure (255) is an error. Exiting the shell prints a short closing line instead of "Error: failed to connect to SSH: ... exit status 127".
- `rm` now shuts down a live VM gracefully (SIGTERM, then SIGKILL as fallback) and always frees the forwarded port instead of aborting when the port does not free within the wait window.

## [0.1.13] — 2026-09-09

### Changed

- Refactored QEMU modules: split cloud-init and lifecycle into smaller files for maintainability.

## [0.1.11] — 2026-09-08

### Changed

- Stopped copying the host `.zshrc` into new VMs during `boite create`.
- Restored the missing embedded default `cmd/qemu/cloudinit.yml` for config fallback.
- Added a creation-time notice when falling back to the default config because `~/.boite.yml` is absent.

### Fixed

- Fixed `cmd/qemu/lifecycle.go` to build against the current QEMU runtime helpers after the native refactor.
- Removed obsolete `.tmux.conf_local` and `.zshrc_local` files from the repo root.
- Configuration merging now correctly preserves user-provided sections (write_files, packages, runcmd) instead of overwriting defaults
- VM resource settings (cpu, memory, disk) are now properly applied from configuration files at startup
- SSH key injection logic correctly merges user keys with instance-specific keys
- APT source validation prevents Ubuntu-specific sources on Debian 12 VMs

## [Unreleased]

## [0.1.10] — 2026-09-08

### Changed

- Removed Ubuntu-specific APT sources (Docker, netplan.io, open-iscsi) from default cloud-init configuration for Debian 12 compatibility.
- Improved config merging to append user packages/runcmd/users to defaults instead of overwriting.
- Added SSH key merging logic to preserve user-authorized keys while injecting instance-specific keys.

### Fixed

- Added APT source validation to prevent Ubuntu-specific sources on Debian 12 VMs.
- Fixed missing purge command registration causing test failures.
- Fixed SSH key injection logic to properly merge user keys with generated instance keys.

## [0.1.7] — 2026-09-05

### Added

- Charmbracelet Lipgloss integration: styled instance status table, colored banners, cards, and animated braille spinners.
- Cloud-init YAML syntax validation and auto-healing before writing to disk.
- Unit test suite covering all CLI commands, flag parsing, cloud-init generation, and UI rendering.
- GoReleaser release workflow and Homebrew cask publishing.
- Facile catalog integration for `facile install boite`.
- QEMU-based VM runtime replacing Multipass (removed multipassPath(), checkCommand(), ensureVMRunning() functions)
- Removed deprecated cloud-init files and replaced with qemu/cloudinit.go
- Updated README.md to remove Multipass prerequisites section
- Updated PLAN.md and PLAN_QEMU_REFACTOR.md to document migration
- Added filet quality check integration
- Added Go test coverage for all commands

### Fixed

- Cloud-init YAML indentation bug in `.zshrc` provisioning that caused `yaml-cpp: error at line 62, column 1: end of map not found`.
- Removed debug log emission on `boite list` and other read-only commands.
- Deferred cloud-init file generation to `boite create`, avoiding unwanted disk writes during informational commands.
- Fixed misleading "Try with sudo" suggestions on non-permission errors.
- Removed broken `boite update` self-updater adhering to Facile Suite Standard §3.1.
- Fixed `.zshrc_local` paths for user `boite` and safe Starship/mise activation.

## [0.1.6] — 2026-09-04

### Changed

- Default cloud-init user set to `boite` exclusively.
- ASCII banner display on shell entry.

[Unreleased]: https://github.com/FacileStudio/boite/compare/v0.1.12...HEAD
[0.1.12]: https://github.com/FacileStudio/boite/releases/tag/v0.1.12
[0.1.11]: https://github.com/FacileStudio/boite/releases/tag/v0.1.11
[0.1.10]: https://github.com/FacileStudio/boite/releases/tag/v0.1.10
[0.1.9]: https://github.com/FacileStudio/boite/releases/tag/v0.1.9
[0.1.8]: https://github.com/FacileStudio/boite/releases/tag/v0.1.8
[0.1.7]: https://github.com/FacileStudio/boite/releases/tag/v0.1.7
[0.1.6]: https://github.com/FacileStudio/boite/releases/tag/v0.1.6
