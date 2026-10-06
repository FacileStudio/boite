# Plan — Neovim via bob and multi-distro VM images (Alpine, NixOS)

Status: **planned.** Checked against: filet suite gate (limits: 250 lines/file, 8 funcs/file, no inline comments, no comments in function bodies), CLI conventions, shell fingerprint trap (`BuildSSHArgs` single quoted string).

## Goal
Switch Neovim installation in the Debian bake to `bob` for latest Neovim, and introduce multi-distro base image support in Boite so instances can run Debian, Alpine, or NixOS.

## Why (evidence)
- `scripts/bake-provision.sh:16` installs `neovim` through `apt-get`, locking the baked image to Debian 13's packaged Neovim release rather than upstream stable/latest.
- `cmd/qemu/paths.go:13-15` hardcodes a single Debian image (`BaseImageName`, `BaseImageURL`, `BaseImageSHA256`).
- `cmd/qemu/provision.go:54-61` hardcodes `sudo apt-get` for package provisioning, failing on non-Debian distributions like Alpine (`apk`) or NixOS (`nix`).

## Approach
Split into two distinct tracks:
1. **Track A (Bake toolchain):** Replace `apt-get install neovim` with pinned `bob` binary installation (`v4.2.0`) in `scripts/bake-provision.sh`. Run `bob use latest` as user `boite`, add `~/.local/share/bob/nvim-bin` to guest PATH (`.zshrc`, `.zshenv`), and symlink into `/usr/local/bin/nvim`.
2. **Track B (Multi-distro runtime & bake):** Define a `DistroSpec` registry in `cmd/qemu/distro.go` specifying image name, download URL, SHA256 checksum, and package manager command (`apt`, `apk`, `nix`). Add `vm.distro` to `BoiteConfig` and `--distro` to `boite create`. Store `Distro` on `Instance`. Generalize `cmd/qemu/provision.go` package installation by distro manager. Create bake scripts for Alpine (`scripts/bake-alpine.sh`) and NixOS (`scripts/bake-nixos.sh`) meeting Boite's guest contract (`boite` user with passwordless root, SSH on port 22, `BOITECFG` firstboot oneshot, `tiroir` static binary).

## Steps (ordered)

### Track A — Neovim via bob in Debian bake
1. `scripts/bake-provision.sh` — remove `neovim` from `apt-get install` packages. [bake]
2. `scripts/bake-provision.sh` — add `BOB_VERSION=v4.2.0`, download `bob-linux-x86_64.zip` from GitHub releases into `/usr/local/bin/bob` (mode 0755), and run `su - boite -c 'bob use latest'`. [bake]
3. `scripts/bake-provision.sh` — add `$HOME/.local/share/bob/nvim-bin` to PATH in `/home/boite/.zshrc` and `/home/boite/.zshenv`, and symlink `/usr/local/bin/nvim` to `/home/boite/.local/share/bob/nvim-bin/nvim`. [bake]

### Track B — Multi-distro engine & CLI
4. `cmd/qemu/distro.go` — create `DistroSpec` struct, built-in distros table (`debian`, `alpine`, `nixos`), lookup helper `GetDistro(name string)`, and validation. [filet]
5. `cmd/qemu/config.go` — add `Distro string` to `VMConfig` (`yaml:"distro"`), defaulting to `debian`. [filet]
6. `cmd/qemu/instance.go` — add `Distro string` to `Instance` struct (`json:"distro"`). [filet]
7. `cmd/qemu/image.go` — parameterize `EnsureBaseImage(distro string)` and `downloadBaseImage(spec DistroSpec)` to download and cache by distro name (`~/.boite/cache/<distro>.qcow2`). [filet]
8. `cmd/qemu/lifecycle.go` — pass target distro through `prepareInstanceDisk`, store distro in `Instance`, and supply it when rebuilding overlays. [filet]
9. `cmd/qemu/provision.go` — refactor `provisionPackages` to branch on `spec.PkgManager` (`apt`: `sudo apt-get install`, `apk`: `sudo apk add --no-cache`, `nix`: `nix profile install`). [filet]
10. `cmd/create.go` — add `--distro` flag to `boite create` command, forwarding value to `lifecycle.Create`. [filet]

### Track C — Distro base images & firstboot
11. `scripts/firstboot-alpine.sh` & `scripts/bake-alpine.sh` — build minimal Alpine qcow2 via `alpine-make-vm-image`: OpenRC `boite-firstboot` service mounting `BOITECFG`, user `boite`, `doas`/`sudo`, OpenSSH, static `tiroir` binary. [bake]
12. `nix/boite-configuration.nix` & `scripts/bake-nixos.sh` — build minimal NixOS qcow2 via `nixos-rebuild build-image --image-variant qcow` or `make-disk-image.nix`: systemd `boite-firstboot.service` mounting `BOITECFG`, user `boite`, passwordless sudo, OpenSSH, static `tiroir` binary. [bake]

## Files to Modify / New
- `scripts/bake-provision.sh` — remove apt neovim, install bob v4.2.0, run `bob use latest`, update PATH and symlink
- `cmd/qemu/distro.go` — new: `DistroSpec` registry and lookup functions
- `cmd/qemu/config.go` — add `Distro` to `VMConfig`
- `cmd/qemu/instance.go` — add `Distro` to `Instance`
- `cmd/qemu/image.go` — support multi-distro caching and checksumming
- `cmd/qemu/lifecycle.go` — wire distro parameter into disk preparation and instance state
- `cmd/qemu/provision.go` — distro-aware package installation command builder
- `cmd/create.go` — `--distro` CLI flag
- `scripts/bake-alpine.sh` — new: Alpine Linux bake script
- `scripts/firstboot-alpine.sh` — new: OpenRC firstboot script
- `nix/boite-configuration.nix` — new: NixOS declarative base configuration
- `scripts/bake-nixos.sh` — new: NixOS bake script

## Exit criteria
1. `filet check` runs completely clean on all Go files.
2. `make test` (`go test ./...`) passes.
3. Running `scripts/bake-provision.sh` installs `bob v4.2.0`, runs `bob use latest`, and leaves `nvim --version` showing latest upstream Neovim on PATH.
4. `boite create test-deb --distro debian` creates and boots Debian VM with apt provisioning.
5. `boite create test-alp --distro alpine` creates and boots Alpine VM with apk provisioning and functional firstboot handshake.
6. `boite create test-nix --distro nixos` creates and boots NixOS VM with functional firstboot handshake.
7. `boite list` reports distro metadata for each instance.

## Risks / unknown unknowns
- **musl compatibility (Alpine):** Tools compiled against glibc will fail unless musl binaries or `gcompat` are supplied. `tiroir` is a static Go binary and works out of the box.
- **NixOS disk layout & immutability:** NixOS system files are in `/nix/store` and read-only. Firstboot must write user SSH keys to `/home/boite/.ssh/authorized_keys` and the marker to `/var/lib/boite/firstboot.done`, both of which are mutable state locations.
- **Bake reproducibility:** `bob use latest` reaches out to GitHub at bake time. The baked `.qcow2` output is what gets pinned by SHA256 in Boite, so runtime VM creation remains completely reproducible.

## Skip (YAGNI)
- Dynamic runtime kernel switching or custom bootloader menus.
- Custom distro ISO installation installer wizard.
- Converting running instances from one distro to another.
