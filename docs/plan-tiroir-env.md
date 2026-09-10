# Plan — tiroir: env system for boite VMs

Status: proposed. No code written. This is the spec to approve before executing.

## Goal
Give boite VMs a first-class env system: a new Go env tool **tiroir** (local-first default) that VM shells load by default and boite manages from the host, with casier as an opt-in scoped source.

## Why (evidence)
A boite VM today ships with no env at all — secrets like `OPENROUTER_API_KEY`, `GH_TOKEN`, Facile `<TOOL>_TOKEN` cannot reach an agent in the sandbox, so agent workflows (Atelier, etc.) stall right where auth starts. The existing systemd/firstboot bake (`scripts/bake-provision.sh`), the config-disk pipe (`cmd/qemu/firstboot.go` `BuildConfigDisk`), and per-`exec` ssh injection (`cmd/qemu/ssh.go`) already exist and are the correct seams to reuse.

## Approach
Two repos: a new `tiroir` (Go lib + CLI, successor to skatos, **CRUD + `export` only — no `run` prefix command**; binary name `tiroir`) and boite gaining a `boite env` host command surface. The source is `local` by default (values injected at create) with an opt-in `casier` source. In casier mode the **host** holds the scoped token and materializes the project's values into the VM's tiroir store — casier is never installed in the guest, and the guest holds only a bounded snapshot, never a live token. **There is no `sync` command**: env is refreshed automatically at the host→guest boundary — `boite run` and `boite exec` re-materialize managed keys from their sources before the command/shell runs, so entering the VM always has the current env. Manual `boite env set` writes straight into the guest store and **sticks** (authoritative over source refresh). Offline / casier-unreachable, the last materialized snapshot is kept with a warning. Presence is ambient: the baked `.zshrc` / `.bashrc` source `tiroir export` at login (with `tiroir` installed into the base image), and `boite exec` re-applies the env for non-interactive `sh -lc` (which never reads rc).

Checked against: casier conventions, agent-access-facile-apps (agents need own identity), distribute/module-path, filet, porte-machine-tokens, cli catalog auth.

---

## Steps

**Track A — tiroir (`github.com/FacileStudio/tiroir`, new repo)**

1. `tiroir/go.mod` — module `github.com/FacileStudio/tiroir`; minimal deps (`gopkg.in/yaml.v3` + cobra if flags wanted). `[module-path]`
2. `tiroir/lib/env.go` — `EnvStore` type backed by **encrypted-at-rest** storage. Split keyfile: `~/.tiroir` holds the ciphertext blob (`0600`), `~/.tiroir.key` holds a random 32-byte key (`0600`). AEAD (AES-256-GCM) with a version + key-id header so keys can rotate later. Per invoking user via `os.UserHomeDir`. Methods: `Get`, `Set`, `Delete`, `List`, `Export` (decrypts then returns `KEY=value` lines, so the ambient rc sourcing is unchanged). Atomic rewrite on write. *Reason:* defeats passive value-scanners that grep for `KEY=` strings; not a real boundary against a tool that runs `tiroir` itself — defense against lazy scanners only.
3. `tiroir/lib/env_test.go` — `filet`-clean tests for round-trip, perms, empty-key, `0600` enforcement. `[filet]`
4. `tiroir/cmd/tiroir/main.go` — CLI (cobra or stdlib `flag`, whatever the repo chooses): `set/get/list/delete/export`. **No `run`.** `[filet]`
5. ~~`tiroir/cmd/skatos_compat`~~ — SKIP, confirmed dead. Do not create.
6. Release tiroir (tag on `master`, goreleaser, matches boite's flow).

**Track B — boite host surface**

7. `boite/go.mod` — add `github.com/FacileStudio/tiroir` as a dependency; `go.mod` `require` + `replace` pinned to the local path during dev, `[distribute]` `github:FacileStudio/tiroir#vX` for released. `[module-path]`
8. `boite/cmd/qemu/config.go` — extend `BoiteConfig` with an `env:` block (see Config below) + an `EnvConfig` / `EnvSource` type (`local` | `casier`).
9. `boite/cmd/qemu/config_test.go` — parse tests for the new block. `[filet]`
10. `boite/cmd/env.go` — new `boite env` cobra command tree: `list/set/get/delete`. Host-side; `set/get/delete` CRUD the VM store over ssh (writes are authoritative and stick). No `sync` sub-command — freshness is a boundary-trigger, not a command.
11. `boite/cmd/qemu/ssh.go` — add a helper that paints tiroir env into the command env for non-interactive exec: `ssh … 'env $(tiroir export); <cmd>'` or fetch-then-`env`. Name it `WithEnv(inst, args)`. `[filet]`
12. `boite/cmd/qemu/firstboot.go` — extend `BuildConfigDisk` to drop a `tiroir.env` (the *resolved values only* — never a casier token) onto the vfat disk alongside `authorized_keys`, initializing the store at VM creation; extend the firstboot oneshot (bake side, Track C) to copy it into `~/.tiroir`. Guard with the existing "finish if no vfat" logic.

**Track C — baked base image + guest**

13. `boite/scripts/bake-provision.sh` — `apt` or copy-install the `tiroir` binary; prepend `eval "$(tiroir export)"` to `/home/boite/.zshrc` and create/append `/home/boite/.bashrc` with the same; keep everything `chown boite:boite`. `[filet]`
14. `boite/scripts/boite/firstboot.sh` (the oneshot laid by bake) — after mounting BOITECFG, if `tiroir.env` present copy it to `/home/boite/.tiroir`, `chmod 0600`, `chown boite:boite`. Mirrors the existing `authorized_keys` write (lines 178-180 + `mkdir -p`).
15. `boite/`base repin + SHA256 in `cmd/qemu/instance.go` — rebuild the baked image and repin `BaseImageName/URL/SHA256`.

**Track D — docs / conventions**

16. Retire skatos: mark `~/.mycelium/memory/tools/skatos.md` superseded (point to tiroir), repoint `~/.agents/skills/skatos` → tiroir or mark superseded.
17. Update `README.md` (Usage + Features) and add an `env:` section to the `Configuration` block.

---

## Config shape (`~/.boite.yml`)

```yaml
env:
  source: local            # local | casier
  local:
    vars:          # literal values baked at create
      LOG_LEVEL: debug
      # resolved by name from the host store at create/entry; list, not literal
      - OPENROUTER_API_KEY
  casier:
    project: my-org
    environment: dev       # trust tier ≈ casier project; unset ⇒ scratch (least privilege)
    token_ref: CASIER_TOKEN   # read-only, scoped casier_… token; never the human's master
```

Rule: `local.vars` name-keys are **resolved from the host's secret store (casier / skatos / host tiroir) at create and entry time**; literal inline values are only for non-secret config (LOG_LEVEL, etc.). The VM's store never holds a host master.

Precedence: **manual `boite env set` beats source refresh.** On entry, boite re-applies managed keys from sources, then re-applies manually-set values on top, so a per-VM override sticks across entries. (Settled 2026-09-10; flip to "source wins" if preferred.)

- **Surface-change (2026-09-10):** EnvStore encrypts at rest via a split keyfile — `~/.tiroir` (ciphertext) + `~/.tiroir.key` (random 32-byte key), both `0600`. **Option 1** chosen over a machine-bound key (no keychain, no password; Option 1 keeps a visible `.key` but is simplest and dependency-light). ADVISORY: keeps a passive `KEY=` value-scanner from triggering; an adversary who can run `tiroir` still reads everything. Not security, just an anti-lazy-scanner layer.
- **Store leak guarantees are stronger:** the config-disk payload (`tiroir.env`) materializes the *encrypted* store, so the guest never sees plaintext env on disk or on the wire — no host token ever leaves in cleartext.

Store path: `~/.tiroir` + `~/.tiroir.key`, both `0600`, per invoking user. Unless the repo chooses XDG, this is fixed.

---

## Files to modify / new

- `tiroir/` (new repo): `go.mod`, `lib/env.go`, `lib/env_test.go`, `cmd/tiroir/main.go`
- `boite/go.mod` — dep + replace
- `boite/cmd/qemu/config.go` — `env:` block
- `boite/cmd/qemu/config_test.go` — tests
- `boite/cmd/env.go` — `boite env` surface
- `boite/cmd/qemu/ssh.go` — env injection for exec
- `boite/cmd/qemu/firstboot.go` — tiroir disk payload
- `boite/scripts/bake-provision.sh` — tiroir install + rc lines + firstboot read
- `boite/scripts/bake-image.sh` — repin (if changed)
- `boite/cmd/qemu/instance.go` — new pinned image
- docs: `README.md`, `tools/skatos.md` (supersede), skatos skill

---

## Exit criteria (verifiable "done")

- `boite create name` → the VM's interactive zsh and `boite exec` both expose `tiroir list` showing the `local.vars` + inline `vars`; `env list` matches.
- `boite env set FOO bar` writes into `~/.tiroir` (ciphertext, `0600`, owned `boite`, key in `~/.tiroir.key` `0600`), appears on the next `boite run` / `exec`, and **remains** after re-entry (manual set is authoritative over source refresh).
- `boite env sync` is **not a command** — instead, re-entering a casier-backed VM (`boite run` / `boite exec`) refreshes managed keys from the scoped token, and the previous snapshot is kept with a warning when casier is unreachable.
- A store-leak module (a VM bound only to scratch creds) cannot reach a master token.
- `boite` builds, `filet check` clean, `tiroir` build + lib tests green (encryption round-trip, perms, empty-key).

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