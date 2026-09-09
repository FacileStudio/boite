# Changelog

All notable changes to this project are documented here. The format is
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html). While on
`0.x`, a breaking change bumps the minor.

## [Unreleased]

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
