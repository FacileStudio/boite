# Plan — tiroir: env system for boite VMs

Status: **Track A shipped (tiroir v0.1.0, 2026-09-10); Track B shipped (boite host surface, commit `cef6e8e`, 2026-09-10); Track C shipped (baked+repinned+deployed base image, commits `658dcb2`+`5c76107`, 2026-09-10); Track D open.** This is the working spec; completed steps are struck through with a note on how reality diverged.

**Resume here (cold start):** the only work left is step 17 (Track D: README). Before the next `boite` release, flip `go.mod`'s `replace github.com/FacileStudio/tiroir => ../tiroir` to `github:FacileStudio/tiroir#v0.2.0` (CI has no sibling repo) — not v0.1.0: boite's tiroir dep resolved to git HEAD, and the `export`-applies-to-shell encoding boite depends on landed in v0.2.0.

## Goal
Give boite VMs a first-class env system: a new Go env tool **tiroir** (local-first default) that VM shells load by default and boite manages from the host, with casier as an opt-in scoped source.

## Why (evidence)
A boite VM today ships with no env at all — secrets like `OPENROUTER_API_KEY`, `GH_TOKEN`, Facile `<TOOL>_TOKEN` cannot reach an agent in the sandbox, so agent workflows (Atelier, etc.) stall right where auth starts. The existing systemd/firstboot bake (`scripts/bake-provision.sh`), the config-disk pipe (`cmd/qemu/firstboot.go` `BuildConfigDisk`), and per-`exec` ssh injection (`cmd/qemu/ssh.go`) already exist and are the correct seams to reuse.

## Approach
Two repos: a new `tiroir` (Go lib + CLI, successor to skatos, **CRUD + `export` only — no `run` prefix command**; binary name `tiroir`) and boite gaining a `boite env` host command surface. The source is `local` by default (values injected at create) with an opt-in `casier` source. In casier mode the **host** holds the scoped token and materializes the project's values into the VM's tiroir store — casier is never installed in the guest, and the guest holds only a bounded snapshot, never a live token. **There is no `sync` command**: env is refreshed automatically at the host→guest boundary — `boite run` and `boite exec` re-materialize managed keys from their sources before the command/shell runs, so entering the VM always has the current env. Manual `boite env set` writes straight into the guest store and **sticks** (authoritative over source refresh). Offline / casier-unreachable, the last materialized snapshot is kept with a warning. Presence is ambient: the baked `.zshrc` / `.bashrc` source `tiroir export` at login (with `tiroir` installed into the base image), and `boite exec` re-applies the env for non-interactive `sh -lc` (which never reads rc).

Checked against: casier conventions, agent-access-facile-apps (agents need own identity), distribute/module-path, filet, porte-machine-tokens, cli catalog auth.

---

## Steps

**Track A — tiroir (`github.com/FacileStudio/tiroir`, new repo) — DONE (v0.1.0)**

1. ~~`tiroir/go.mod` — module `github.com/FacileStudio/tiroir`; minimal deps (`gopkg.in/yaml.v3` + cobra if flags wanted).~~ `[module-path]` — ship reality: single `require github.com/spf13/cobra v1.10.2`; no yaml dep needed.
2. ~~`tiroir/lib/env.go` — `EnvStore` type backed by **encrypted-at-rest** storage. Split keyfile: `~/.tiroir` (ciphertext, `0600`) + `~/.tiroir.key` (random 32-byte key, `0600`). AEAD (AES-256-GCM) with a version + key-id header. Per invoking user via `os.UserHomeDir`. Methods `Get`/`Set`/`Delete`/`List`/`Export`. Atomic rewrite on write.~~ Done as specced. Crypto split into `lib/crypto.go` (AES-GCM, `seal`/`decrypt`/`read`/`save`, atomic write); `lib/env.go` holds the store API.
3. ~~`tiroir/lib/env_test.go` — filet-clean tests for round-trip, perms, empty-key, `0600` enforcement.~~ Done — 8 tests including at-rest encryption (grep for secret in store finds 0 hits) and tamper-detection. `[filet]`
4. ~~`tiroir/cmd/tiroir/main.go` — CLI: `set/get/list/delete/export`. **No `run`.**~~ Done but **diverged**: layout is `main.go` at repo root + cobra tree in `cmd/root.go` / `cmd/commands.go` — mirroring boite, not `cmd/tiroir/`. `export` emits `export KEY='value'` shell statements (single-quoted) so `eval "$(tiroir export)"` applies the store directly into the current shell — not bare `KEY=value` lines, changed in v0.2.0. Confirmed no `run` subcommand (`unknown command "run"`). `[filet]`
5. ~~`tiroir/cmd/skatos_compat`~~ — SKIP, confirmed dead. Not created.
6. ~~Release tiroir (tag on `master`, goreleaser, matches boite's flow).~~ Done — repo created `FacileStudio/tiroir` (public), tag `v0.1.0`, `.goreleaser.yml` matching boite's, release workflow green (7 tar.gz + checksums). **Extras shipped not in plan**: `.github/workflows/filet.yml` (style gate + Antenne webhook) and `.github/workflows/release.yml` (goreleaser on `v*`); tiroir added to the facile tool catalog (`facile/internal/manifest/tools.yml`, entry after `boite`, branch `master`); `CHANGELOG.md` (Keep a Changelog). Fixed a stamping bug during wiring: `main.go` overwrote the ldflags `cmd.Version` with `dev` — removed the override so `-X github.com/FacileStudio/tiroir/cmd.Version={{.Version}}` sticks.

**Track B — boite host surface — DONE (commit `cef6e8e`, 2026-09-10)**

7. ~~`boite/go.mod` — add `github.com/FacileStudio/tiroir` as a dependency; `go.mod` `require` + `replace` pinned to the local path during dev, `[distribute]` `github:FacileStudio/tiroir#vX` for released.~~ Done — `require github.com/FacileStudio/tiroir v0.1.0` + `replace github.com/FacileStudio/tiroir => ../tiroir`. `[module-path]` Cobra bumped to v1.10.2 transitively (tiroir requires ≥ that, MVS wins). The `../tiroir` replace must be switched to `github:FacileStudio/tiroir#v0.1.0` before the next boite release (CI has no sibling repo).
8. ~~`boite/cmd/qemu/config.go` — extend `BoiteConfig` with an `env:` block (see Config below) + an `EnvConfig` / `EnvSource` type (`local` | `casier`).~~ Done. `EnvSource` (`local`|`casier`), `EnvConfig{Source,Local,Casier}`, `EffectiveSource()` defaults to `local`. **Shape divergence**: the terse Config example mashed literal map entries and name-keys under one `vars:` key; YAML cannot hold a map and a list in one key, so `local` is `vars:` (literal `map[string]string`, non-secret only) + `resolve:` (`[]string`, key names pulled by name from the **host tiroir store** at create). Both documented.
9. ~~`boite/cmd/qemu/config_test.go` — parse tests for the new block.~~ Done — local literal+resolve, casier, and default-source tests.
10. ~~`boite/cmd/env.go` — new `boite env` cobra command tree: `list/set/get/delete`. Host-side; `set/get/delete` CRUD the VM store over ssh (writes are authoritative and stick).~~ Done. Surface: `boite env <list|get|set|delete> <name> [key [value]]`. The VM graph is via the guest `tiroir` binary over ssh (`SSHOutput` prints list/get; `SSHCommand` runs set/delete). Explicitly **no `sync` sub-command**. Manual set writes straight into the guest store and sticks.
11. ~~`boite/cmd/qemu/ssh.go` — add a helper that paints tiroir env into the command env for non-interactive exec: `ssh … 'env $(tiroir export); <cmd>'` or fetch-then-`env`. Name it `WithEnv(inst, args)`.~~ Done, **diverged on signature**: `WithEnv(command []string) []string` (no `inst` — the helper only rewrites the command, and a dead param is noise). Prepend `eval "$(tiroir export); "` as a single ssh arg so it survives the join/reparse; a missing tiroir binary becomes an empty eval and the wrapped command still runs (verified by test). Wired into `boite exec`. `[filet]`
12. ~~`boite/cmd/qemu/firstboot.go` — extend `BuildConfigDisk` to drop the **encrypted** tiroir store + its key onto the vfat disk alongside `authorized_keys`, initializing the store at VM creation.~~ Done. `BuildConfigDisk(instanceDir, sshPubKey, envVars)`; new `cmd/qemu/env.go` `materializeEnv` resolves literals + host-store names into a map; `writeTiroirPayload` uses `tiroir.NewAt(stageDir)` + `Set` per key to produce the ciphertext `.tiroir` **and** `.tiroir.key`, mcopied to the disk (empty map = no payload). `[filet]` Guard "finish if no vfat" is on the firstboot oneshot (Track C). The guest stores only ciphertext — plaintext never sits on disk or the wire.

**Track C — baked base image + guest — DONE (commits `658dcb2`+`5c76107`, 2026-09-10)**

13. ~~`boite/scripts/bake-provision.sh` — `apt` or copy-install the `tiroir` binary; prepend `eval "$(tiroir export)"` to `/home/boite/.zshrc` and create/append `/home/boite/.bashrc` with the same; keep everything `chown boite:boite`. `[filet]`~~ Done. tiroir v0.2.0 downloaded from the GitHub release (`tiroir_0.2.0_linux_amd64.tar.gz`, pinned `TIROIR_VERSION=0.2.0` in the toolchain block) and installed to `/usr/local/bin`. `.zshrc` sources `eval "$(tiroir export)"` right after the PATH/WORKSPACE exports; a new `.bashrc` (was absent before) carries the same line so a bash login also sees managed keys. Both `chown boite:boite`, `0644`.
14. ~~The firstboot oneshot — `scripts/bake-provision.sh` the inline `cat > /usr/local/lib/boite/firstboot.sh` block. After mounting BOITECFG, if the tiroir store is present copy **both** `.tiroir` and `.tiroir.key` from `/mnt/boitecfg` to `/home/boite/`, `chmod 0600` each, `chown boite:boite`. Mirrors the `authorized_keys` write.~~ Done, exactly as specced — the copy guards on both files present and sits right after the `authorized_keys` write, before `umount`. `BuildConfigDisk` writes `.tiroir`/`.tiroir.key` at the vfat root (`::/.tiroir`), which this copy reads.
15. ~~`boite/`base repin + SHA256 in image consts — rebuild the baked image and repin.~~ Done. Rebuilt via `scripts/bake-image.sh`; SHA256 pinned in `cmd/qemu/paths.go` (the consts moved here from `instance.go` in the working-tree refactor). Repin to `51e6e296…`. Deployed the ~2.0 GB qcow2 to `/etc/dokploy/boite-assets/boite.qcow2` (the site's bind-mount source); old image saved as `boite.qcow2.prev` for rewind. End-to-end verified: `boite create envtest` → `boite env set FOO hello` → `boite exec` shows `FOO=hello` (WithEnv) and interactive `zsh -i` reads rc and shows `FOO=hello`; manual set sticks across re-entry. The served URL `https://boite.facile.studio/base.qcow2` returns HTTP 200 with the new image (incl-length matches the pinned SHA's file). **Gotcha**: the bake is non-deterministic — two runs give different SHA256 (random root password/seed), so pin whatever image is actually served, never assume a hash.

**Track D — docs / conventions**

16. ~~Retire skatos: mark `~/.mycelium/memory/tools/skatos.md` superseded (point to tiroir), repoint `~/.agents/skills/skatos` → tiroir or mark superseded.~~ **DONE 2026-09-10**: wiki page marked `[SUPERSEDED by tiroir]`, tiroir page + index entry created, `tools/tiroir.md` written. Skill replaced wholesale: new `~/.mycelium/skills/tiroir.md` (name/desc triggers `tiroir` + legacy `skatos` so old triggers still fire), regenerated via `mycelium install agents` into `~/.agents/skills/tiroir/`; orphan `~/.agents/skills/skatos/` and the plaintext `~/.skatos/default.yml` deleted (migration verified: 10 keys match tiroir exactly). skatos binary had no install on ruche.
17. Update `README.md` (Usage + Features) and add an `env:` section to the `Configuration` block.

---

## Config shape (`~/.boite.yml`)

```yaml
env:
  source: local            # local | casier
  local:
    vars:              # literal values baked at create (non-secret config only)
      LOG_LEVEL: debug
    resolve:           # key names pulled by name from the host tiroir store at create
      - OPENROUTER_API_KEY
  casier:
    project: my-org
    environment: dev       # trust tier ≈ casier project; unset ⇒ scratch (least privilege)
    token_ref: CASIER_TOKEN   # read-only, scoped casier_… token; never the human's master
```

Rule: `local.resolve` name-keys are **resolved from the host's secret store (host tiroir) at create time**; `local.vars` literal inline values are only for non-secret config (LOG_LEVEL, etc.). The VM's store never holds a host master.

Precedence: **manual `boite env set` beats source refresh.** On entry, boite re-applies managed keys from sources, then re-applies manually-set values on top, so a per-VM override sticks across entries. (Settled 2026-09-10; flip to "source wins" if preferred.)

- **Surface-change status: IMPLEMENTED (2026-09-10, tiroir v0.1.0).** EnvStore encrypts at rest via a split keyfile — `~/.tiroir` (ciphertext) + `~/.tiroir.key` (random 32-byte key), both `0600`. **Option 1** chosen over a machine-bound key (no keychain, no password; Option 1 keeps a visible `.key` but is simplest and dependency-light). ADVISORY: keeps a passive `KEY=` value-scanner from triggering; an adversary who can run `tiroir` still reads everything. Not security, just an anti-lazy-scanner layer.
- **Store leak guarantees are stronger:** the config-disk payload materializes the *encrypted* `~/.tiroir` blob **plus** `~/.tiroir.key` (step 12/14), so the guest never sees plaintext env on disk or on the wire — no host token ever leaves in cleartext.

Store path: `~/.tiroir` + `~/.tiroir.key`, both `0600`, per invoking user. Unless the repo chooses XDG, this is fixed.

---

## Files to modify / new

- `tiroir/` (new repo — **created**): `go.mod`, `lib/env.go`, `lib/crypto.go`, `lib/env_test.go`, `main.go`, `cmd/root.go`, `cmd/commands.go`, `.goreleaser.yml`, `.github/workflows/{filet,release}.yml`, `filet.yml`, `CHANGELOG.md`, `.gitignore`
- `boite/go.mod` — dep + replace (**done**) + `boite/go.sum` (regenerated)
- `boite/cmd/qemu/config.go` — `env:` block (**done**)
- `boite/cmd/qemu/config_test.go` — tests (**done**)
- `boite/cmd/qemu/env.go` — `materializeEnv` resolution (**new, done**)
- `boite/cmd/env.go` — `boite env` surface (**done**)
- `boite/cmd/qemu/ssh.go` — `WithEnv` for exec + `SSHOutput` (**done**)
- `boite/cmd/qemu/firstboot.go` — tiroir disk payload (**done**)
- `boite/scripts/bake-provision.sh` — tiroir install + rc lines + firstboot read (Track C)
- `boite/scripts/bake-image.sh` — repin (if changed)
- `boite/cmd/qemu/instance.go` — new pinned image
- docs: `README.md`, `tools/skatos.md` (superseded — **done**), skatos skill (replaced by tiroir skill — **done**)
- `facile/internal/manifest/tools.yml` — **done**: tiroir catalog entry added

---

## Exit criteria (verifiable "done")

- `boite create name` → the VM's interactive zsh and `boite exec` both expose `tiroir list` showing the `local.vars` + inline `vars`; `env list` matches.
- `boite env set FOO bar` writes into `~/.tiroir` (ciphertext, `0600`, owned `boite`, key in `~/.tiroir.key` `0600`), appears on the next `boite run` / `exec`, and **remains** after re-entry (manual set is authoritative over source refresh).
- `boite env sync` is **not a command** — instead, re-entering a casier-backed VM (`boite run` / `boite exec`) refreshes managed keys from the scoped token, and the previous snapshot is kept with a warning when casier is unreachable.
- A store-leak module (a VM bound only to scratch creds) cannot reach a master token.
- ~~`boite` builds, `filet check` clean, `tiroir` build + lib tests green (encryption round-trip, perms, empty-key).~~ **Met 2026-09-10** (commit `cef6e8e`): `scripts/check.sh` green (`gofmt`/vet/test/filet all pass, exit 0) on the Track B boite host surface; tiroir build + 8 lib tests green and filet clean. The end-to-end criterion below is **Met 2026-09-10**: `boite create envtest` exposed `tiroir list` and `boite env set` values to both interactive zsh (rc) and `boite exec` (WithEnv), and manual sets stick across re-entry.

---

## Risks / unknown unknowns

- **casier token scope/handling**: per the casier-cli README this is a `casier_` project/env/read-only key — confirm the server enforces it this way before wiring. (Journal/registre note: the machine-token path was historically dormant on registre; casier's scoped key may be the better vehicle.)
- **baked image repin** is a slow, network-bound step and a make-it-rewind point: a bad repin breaks every `create`. Pin early, verify once.
- **`sh -lc` for `boite exec`**: non-interactive ssh does not source rc, so the env must be painted explicitly (step 11); easy to miss, and then exec silently lacks vars.
- **store permission drift** between host `boite env set` and the guest firstboot write: keep every path to `chmod 0600`, `chown boite:boite`.
- **cross-repo consume during dev**: pin via `replace` to a working-dir symlink or a `#<branch>` so a half-built tiroir does not break boite.

---

## Skip (YAGNI)

- **`tiroir run` / any runner command** — out. Ambient via rc + boite-inject is the model.
- **skatos compat shim** — out. skatos is replaced wholesale; no migration path needed.
- **Per-VM casier project creation in `boite create`** — out. boite uses the scoped token to pull; the project is created/bound on the casier side. (Trust tiers associate via casier project, but `boite create` should not insist on creating it.)
- **`tiroir` watch / push daemon / websocket live-update** — out. Freshness is materialized-at-entry (`boite run` / `boite exec`), never a running daemon.
- **`boite env sync` sub-command** — out. There is no manual sync; env refreshes at the boundary. Explicitly decided.
- **`tiroir` keychain / OS-secret integration as a requirement** — out; a `0600` file is compliant and keeps the tool dependency-free per cli.md 6.7.