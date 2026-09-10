# Boite: baked base image with virt-builder debian-13

Implémentation plan (cold-start handoff — a fresh worker needs no prior conversation to execute this).

Status: APPROVED — ready to execute. Decisions locked: boite user + SSH host keys baked in, cloud-init purged in the bake, old `state.json` format removed with no legacy read.

## Goal

Bake boite's base image from `virt-builder debian-13`, with user, SSH identity, toolchain and dotfiles preinstalled. `create` becomes copy-overlay + boot + wait for firstboot. Cloud-init leaves the code base and the rewritten config surface entirely.

## Why (evidence) — one line

`cmd/qemu/cloudinit.go` renders a full cloud-init user-data tree, then scrapes the QEMU console (`cmd/qemu/cloudinit_diag.go`) to explain `runcmd` failures — a whole diagnostic layer for provisioning that is identical across every instance, because the base image (`debian-13-genericcloud`) is built to be cloud-init-driven. `virt-builder debian-13` is a plain, DHCP-bootable Debian that needs no cloud-init.

## Approach

Keep boite's `download → repin (URL+SHA) → verify` image model; only what is downloaded changes. A repo script `scripts/bake-image.sh` runs `virt-builder debian-13` on this box (virt-builder is installed, template listed) with `--firstboot-command`, `--install`, `--write`, `--ssh-inject` to produce a versioned boite image, then prints its URL+SHA for `cmd/qemu/instance.go`. The bake seeds: `boite` user + NOPASSWD sudo, SSH host-key regeneration (`dpkg-reconfigure openssh-server`), the full package + toolchain + dotfile list, and a `boite-firstboot` systemd oneshot that installs the one genuinely per-instance input (the generated public key) from a tiny config ISO, then disables itself. All `CloudInit*`/merge/render/diag code is deleted.

Hostname: constant `boite` in the bake (drops the old per-instance `instance-id`). Debian kept; Alpine explicitly out (musl breaks the glibc-targeting toolchain — Go, Rust, bun).

## Decisions (locked)

1. **`boite` user + SSH host key via the bake.** The `virt-builder debian-13` template ships no user and no SSH host keys (only sshd). The bake must create `boite` and run `dpkg-reconfigure openssh-server`.
2. **Remove cloud-init in the bake.** `apt purge cloud-init cloud-images` so the mechanism is verifiably gone (`cloud-init status` errors afterward).
3. **Remove the old `state.json` format. No legacy read.** `seed_iso_path` → `config_iso_path`. Existing sandboxes become orphans and drop out of `list`. Accepted; do not add a compat read.

## Steps (ordered)

1. **Minimal bake spike** — `scripts/bake-image.sh`: `virt-builder debian-13` +
   `--firstboot-command "dpkg-reconfigure openssh-server"` + create `boite` user + `--write` a key placeholder.
   Boot the baked image, assert a fresh overlay reaches SSH. [exit: fresh overlay of the unprovisioned baked base boots to SSH with no cloud-init]
2. **Full bake + repin** — extend the bake: package list from `cmd/qemu/cloudinit.yml` (~30 pkgs), toolchain (mise, Go, Rust, bun, skatos), `.zshrc`/`.tmux.conf`, `firstboot.sh` + systemd oneshot, then `apt purge cloud-init cloud-images`. Repin `BaseImageName` → `boite-base-<version>.qcow2`, `BaseImageURL`, `BaseImageSHA256` in `cmd/qemu/instance.go`. [exit: bake runs, repin set, throwaway overlay boots to complete toolchain]
3. **Config ISO replaces the seed** — new `cmd/qemu/firstboot.go`: `BuildConfigISO(instanceDir, pubKey)` writes `authorized_keys` and runs `genisoimage -output config.iso -r config`. Delete `GenerateSeedISO` / `renderCloudInitUserData` / `writeCloudInitFiles` / `runCloudLocalDS`. Rename `Instance.SeedISOPath` → `ConfigISOPath` (+ JSON `seed_iso_path` → `config_iso_path`), `GetSeedISOPath` → `GetConfigISOPath`, `QEMUConfig.SeedISOPath` → `ConfigISOPath`, `helpers.go` `startFinalizeParams` field. [exit: `make build` clean; an instance mounts `config.iso` and lands `authorized_keys` after boot]
4. **Firstboot wait replaces cloud-init wait** — `cmd/qemu/lifecycle.go`: replace the `WaitForCloudInit` call in `startAndFinalizeInstance` with `WaitForFirstboot(inst)` — reuse the SSH poll-loop shape, wait on `test -f /var/lib/boite/firstboot.done`, keep the progress ticks. Delete `WaitForCloudInit`, `cloudInitStatus` / `cloudInitState` / `isFatalCloudInitState` / `cloudInitLiveLog`, `anyPasswordInUsers`, and all of `cmd/qemu/cloudinit_diag.go`. `ProgressPhase("Building cloud-init seed ISO")` → `"Building config ISO"`. [exit: `create` boots a fresh instance to a provisioned shell; `grep -ri cloud-init cmd/` empty]
5. **Config surface sheds `cloud_init`** — `cmd/qemu/config.go`: delete `CloudInitConfig`, `WriteFileConfig` / `APTConfig` / `APTSource` / `UserConfig`, `MergeCloudInitConfig`, merge helpers (`mergePackages` / `mergeRuncmd` / `mergeUsers` / `mergeWriteFiles`), `findAuthorizedKeys`, `defaultCloudInitConfig`, and the `//go:embed cloudinit.yml` + `DefaultCloudInitYAML`. Drop `cloud_init` from `BoiteConfig`; `cmd/root.go` `createDefaultConfig` writes `vm` + `workspace` only. [exit: `go build` clean, fresh default config has no `cloud_init:` key]
6. **Rename sweep** — delete `cmd/qemu/cloudinit.yml`; update `README.md` (drop the cloud-init install line and the "merged with `cloudinit.yml`" section), `SUMMARY_OF_FIXES.md`, `tmp/real_boite.yml`; remove `cloud-init-debug.yml` from `.gitignore`; replace `cmd/commands_test.go` `TestDefaultCloudInitIsValidYAML` and `cmd/utils_test.go` `TestDefaultCloudInit` with a `TestConfigISO`; in `scripts/check.sh` swap the runtime host dep `cloud-localds` → `genisoimage`. [exit: `grep -ri cloud-init -l` over the tree returns only `CHANGELOG.md` / git history]
7. **Gate** — `sh scripts/check.sh` (gofmt, `go vet`, `go test`, `filet check .`) green; one manual `create` on a machine without `cloud-localds`, one against the pre-baked cache. [exit: check.sh clean end-to-end]

## Convention flags

Checked against: `[filet]` (limits in `filet.yml`), `[cli-standard]`, `[rename-sweep]` (`suite-rename-sweeps.md`), `[proving-a-refactor-changed-nothing]`, `[no-slop]`, `[project-architecture]`.

Not applicable: `[migrations]` (no DB), `[auth/porte]` (no web auth), `[muse]` (no Svelte), `[module-path]` (module stays `github.com/FacileStudio/boite` — change nothing there).

## Files to Modify / New / Delete

- `scripts/bake-image.sh` — **new**: `virt-builder` bake + repin output
- `scripts/firstboot.sh` — **new**: oneshot installed into the bake
- `cmd/qemu/firstboot.go` — **new**: `BuildConfigISO` + `WaitForFirstboot`
- `cmd/qemu/instance.go` — repin; `SeedISOPath` → `ConfigISOPath`
- `cmd/qemu/lifecycle.go`, `cmd/qemu/qemu.go`, `cmd/qemu/helpers.go` — ISO name / wait reshuffle
- `cmd/qemu/config.go`, `cmd/root.go` — drop `cloud_init`
- `cmd/commands_test.go`, `cmd/utils_test.go` — cloud-init tests → `TestConfigISO`
- `scripts/check.sh` — host dep `cloud-localds` → `genisoimage`
- `README.md`, `SUMMARY_OF_FIXES.md`, `tmp/real_boite.yml`, `.gitignore`
- **delete**: `cmd/qemu/cloudinit.go`, `cmd/qemu/cloudinit_render.go`, `cmd/qemu/cloudinit_diag.go`, `cmd/qemu/cloudinit_test.go`, `cmd/qemu/cloudinit.yml`

## Exit criteria

- `filet check .` clean and `sh scripts/check.sh` green end-to-end.
- `grep -ri cloud-init cmd/ README.md scripts/` returns only historical references.
- On a machine without `cloud-localds`, `boite create` boots a fresh instance to a provisioned shell using only a baked image + config ISO.
- `state.json` writes `config_iso_path` (no `seed_iso_path`).

## Risks / unknown unknowns

- **SSH identity bootstrap** — the template has no SSH host keys and no user; the bake must run `dpkg-reconfigure openssh-server` and create `boite`, or there is no handshake / no login. This is step 1's job; verify before step 2.
- **cloud-init presence in the template** — the notes call it "minimal", but cloud-init may still be installed; the bake purges it (decision 2).
- **First-boot race** — confirm the oneshot runs and writes the marker before `WaitForFirstboot`'s first poll on a cold boot; test the marker path explicitly.
- **`state.json` orphans** — old sandboxes (with `seed_iso_path`) drop out of `list`. Accepted by decision 3; no legacy read.

## Skip (YAGNI)

- **Per-instance hostname** — constant `boite` in the bake.
- **Alpine / another OS** — musl breaks glibc-targeting toolchain; Debian kept.
- **Config-drivable `packages` / `runcmd` mutation** — `~/.boite.yml` keeps `vm:` + `workspace:` only.
- **Runtime baking** — boite never shells `virt-builder`; build once, repin, download.
- **Keeping `cloud-localds`** — replaced by `genisoimage`.
- **Rewriting historical `CHANGELOG.md` entries** — add new release notes only.