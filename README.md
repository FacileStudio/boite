# Boite - Development Sandbox Manager

CLI tool to create, manage, and clean up virtual machine-based development sandboxes for general development.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/FacileStudio/boite/main/install.sh | bash
```

Installs to `~/.local/bin` via [facile](https://github.com/FacileStudio/facile), the suite installer. Pass `--bin-dir <dir>` to change that, `--source` to build from source, `--no-skill` to skip AI agent skill registration.

Already have `facile`:

```sh
facile install boite
```

### Build from source

```sh
go build -o boite .
./boite install
```

## Prerequisites

QEMU, `mkfs.fat` and `mcopy` are required for VM-based sandboxes. The default
`open` network also needs [passt](https://passt.top) (Linux only); see
[Egress](#egress).

- Ubuntu/Debian: `sudo apt install qemu-kvm dosfstools mtools passt`
- See [QEMU website](https://qemu.org) for other platforms.

## Usage

```sh
# Create a new sandbox VM with current directory mounted at /workspace
boite create my-project

# Open an interactive shell
boite run my-project

# Copy /workspace from the VM back into the current directory
boite sync my-project

# Execute a command inside the sandbox VM
boite exec my-project ls -la

# List sandboxes
boite list

# Stop a sandbox (preserves data)
boite stop my-project

# Start a stopped sandbox
boite start my-project

# Remove a sandbox permanently
boite rm my-project
```

## Commands

| Command | Description |
|---------|-------------|
| `create <name>` | Create a new sandbox VM |
| `run <name>` | Open an interactive shell (syncs the current directory into /workspace when workspace sync is enabled) |
| `run --no-workspace <name>` | Open the shell without syncing the current directory, even if sync is enabled |
| `exec <name> [cmd...]` | Execute command in a running sandbox VM |
| `sync <name>` | Copy /workspace from the sandbox back into the current directory |
| `list` | List all sandboxes |
| `start <name>` | Start sandbox VM |
| `stop <name>` | Stop sandbox VM (preserves data) |
| `rm <name>` | Destroy sandbox VM permanently |
| `snapshot <name> <tag>` | Snapshot a stopped sandbox's overlay disk |
| `rollback <name> <tag>` | Roll a stopped sandbox's overlay back to a snapshot (discards current state) |
| `snapshots <name>` | List the qcow2 snapshots on a sandbox's overlay |
| `install` | Install boite to system PATH |

## Worker (self-hosted control API)

`boite-worker` is a host daemon that drives the same QEMU lifecycle as the CLI behind a small HTTP API on a Unix socket. A web tier (boite-web, Bulle) calls it; users never reach it directly. See [docs/arch.md](docs/arch.md) for why it is a separate process.

Prerequisites on the host:

- `sudo apt install qemu-kvm dosfstools mtools genisoimage passt`
- `/dev/kvm` readable by the worker (it runs as root)

Build and install:

```sh
go build -o boite-worker ./cmd/boite-worker
sudo install boite-worker /usr/local/bin/
sudo install -m 0644 scripts/boite-worker.service /etc/systemd/system/
sudo install -d -m 0700 /etc/boite
sudoedit /etc/boite/worker.env
sudo systemctl enable --now boite-worker
```

`/etc/boite/worker.env` holds `KEY=value` lines (see below). The unit sets `StateDirectory=boite` (state in `/var/lib/boite`), `RuntimeDirectory=boite` (socket in `/run/boite`) and `KillMode=process`, so stopping or redeploying the worker leaves running boxes alone.

| Variable | Default | Meaning |
|----------|---------|---------|
| `BOITE_WORKER_TOKEN` | none, required | Bearer token every request must carry. The worker refuses to start without it. |
| `BOITE_WORKER_SOCK` | `/run/boite/worker.sock` | Unix socket path. |
| `BOITE_WORKER_SOCK_MODE` | `0600` | Socket permissions as an octal string. The owner needs read and write, and no bit may open it to other users. The worker refuses to start on a bad value. |
| `BOITE_WORKER_SOCK_GROUP` | unset | Group name or numeric gid. The socket is chowned to `root:<group>`. Needs a mode with group read and write, such as `0660`. A client in that group (for example uid 65532, the distroless nonroot user, with `group_add`) can then connect. |
| `BOITE_SSH_BIND` | `127.0.0.1` | Address each box's forwarded SSH port binds to. Use a private address such as a Docker bridge gateway; never `0.0.0.0` on a public host. The CLI also honors it. |
| `BOITE_WORKER_HOST` | value of `BOITE_SSH_BIND` | Host reported to clients in connect info. Set it when the bind is a wildcard. |
| `BOITE_CONTROL_PUBKEY` | unset | One public key added to every new box beside the user's keys, so the caller can SSH in as the control plane. |
| `BOITE_DIR` | `~/.boite` | State root (instances and image cache). The unit sets `/var/lib/boite`. |

Every request needs `Authorization: Bearer <BOITE_WORKER_TOKEN>`. Errors are `{"error": "...", "code": "..."}` with codes `unauthorized`, `invalid`, `not_found`, `exists`, `not_stopped`, `busy`, `not_ready`, `internal`.

| Method and path | Body | Success |
|-----------------|------|---------|
| `POST /boxes` | `{"name","owner","cpus","memory","disk","ssh_key","net"}` (`name` and `owner` required; defaults 2, `2G`, `20G`, `open`) | 202 box with `status` `creating`. The box is created in the background, one at a time. 409 `exists` if the name is taken, including by a box still creating or in error. |
| `GET /boxes[?owner=]` | | 200 `[box]` |
| `GET /boxes/{id}` | | 200 box |
| `POST /boxes/{id}/start` | | 200 box |
| `POST /boxes/{id}/stop` | | 200 box |
| `POST /boxes/{id}/rm` | | 204. Works in every status. On a creating box it cancels a queued create at once, or waits for the running create to end and removes the result. |
| `PUT /boxes/{id}/ssh-key` | `{"ssh_key"}` replaces the user key | 200 box |
| `PUT /boxes/{id}/net` | `{"net"}`, `open`, `offline` or `allowlist`; with `allowlist` an optional `"net_allow": ["host", ...]` replaces the box's extra hosts | 200 box |
| `POST /boxes/{id}/snapshot` | `{"tag"}` | 201 `{"id","tag"}` |
| `POST /boxes/{id}/rollback` | `{"tag"}` | 200 `{"id","tag"}` |
| `GET /boxes/{id}/snapshots` | | 200 `[{"id","tag","created_at"}]` |

A box is `{"id","name","owner","status","cpus","memory","disk","net","net_allow","ssh":{"host","port","user"},"created_at"}`; `net_allow` is present only for `allowlist` boxes with extra hosts; `id` equals `name` (lowercase letters, digits, dashes, max 40). `status` is `running`, `stopped`, `creating` or `error`. A creating box has `ssh.port` 0 until it is up. An `error` box carries an `"error"` string and stays listed until you `rm` it. After a worker restart, boxes left `creating` become `error` with the message `interrupted`. Start, stop, key and snapshot calls on a creating or error box answer 409 `not_ready` (snapshots answer `not_stopped`). Snapshot, rollback and snapshot listing need a stopped box and answer 409 `not_stopped` otherwise. A second call on a box that is mid-operation answers 409 `busy`.

A key set with `PUT /boxes/{id}/ssh-key` is written into a running box at once, and into a stopped box on its next start.

`net` is the box's egress mode, as in [Egress](#egress): `open` runs the box behind its own passt, `offline` sets slirp `restrict=on`, so the guest reaches nothing while SSH through the forwarded port keeps working, and `allowlist` reaches only the built-in registry hosts plus the box's `net_allow` entries through a host-side proxy. `POST /boxes` takes the same `net` and `net_allow`. `net_allow` entries are `host`, `*.host` (subdomains only) or `host:port`, at most 64, and are refused unless `net` is `allowlist`; `PUT` replaces the whole list, so omitting it clears it. Old clients that send only `{"net"}` keep working. The worker stores it in the box's `boite.yml` and `worker.json`. QEMU reads it at boot, so `PUT /boxes/{id}/net` on a running box changes the reported `net` at once but the box keeps its old mode until a stop and start. Boxes created before this field report `open`.

Idle stop lives in the caller. The worker never stops a box on its own: the web tier decides when a box is idle and calls `POST /boxes/{id}/stop`.

```sh
curl --unix-socket /run/boite/worker.sock -H "Authorization: Bearer $BOITE_WORKER_TOKEN" http://worker/boxes
```

## Workspace sync

Workspace sync is an opt-in feature. When enabled, `boite run` copies the
directory you run it from into the sandbox's `/workspace` before opening the
shell, so the VM sees the code you are working on. The copy overwrites
`/workspace` wholesale (any changes made inside the VM are replaced), so push
changes back with `boite sync` before re-running.

The sync is a plain file copy over SSH, not a mount: the VM's `/workspace` is a
bounded snapshot, never a live handle into your host tree. This is what keeps a
compromised guest from reaching your files.

#`allowlist` lets the guest reach a list of hosts and nothing else. boite
starts QEMU with slirp `restrict=on`, which drops every guest-initiated
packet, and adds one `guestfwd` rule that hands each guest TCP connection to
`10.0.2.100:3128` to a boite process on the host. That process is an HTTP
forward proxy: it serves `CONNECT host:port` (HTTPS, no TLS interception) and
`http://` requests only when the host is on the list, resolves the name
itself, and refuses any address that is not public unicast, so an allowed
name cannot reach the host loopback, a LAN or a cloud metadata address. Denied
requests answer 403 and are appended to `egress.log` in the instance
directory. The rules live in QEMU's command line and in files on the host, so
a root user in the guest cannot change them. The guest cannot reach UDP,
ICMP, raw DNS or any IP directly; there is no guest DNS, the proxy resolves
names.

```yaml
vm:
  net: allowlist
  net_allow: [api.example.com, "*.cdn.example.org", git.example.net:9418]
```

The built-in list covers `deb.debian.org`, `security.debian.org`,
`dl-cdn.alpinelinux.org`, `cache.nixos.org`, `registry.npmjs.org`,
`registry.yarnpkg.com`, `pypi.org`, `files.pythonhosted.org`, `crates.io`,
`index.crates.io`, `static.crates.io`, `proxy.golang.org`, `sum.golang.org`,
`github.com`, `api.github.com`, `codeload.github.com` and the GitHub content
hosts `objects.`, `raw.` and `release-assets.githubusercontent.com` (full
list: `cmd/egress/defaults.go`). Matching is on the whole host name,
lowercased: `github.com` does not match `evilgithub.com` or
`sub.github.com`, and `*.example.org` matches `a.example.org` but not
`example.org`. IP literals are refused in the list and in requests. A host
entry covers ports 80 and 443; `host:port` covers that port only.

Tools learn about the proxy from the guest environment. The worker writes
`http_proxy`, `https_proxy` (both cases) and `no_proxy` into `/etc/environment`
and `/etc/profile.d/boite-egress.sh`, and an apt drop-in at
`/etc/apt/apt.conf.d/99boite-egress`, over SSH after each start of an
`allowlist` box, so no image rebake is needed. They are a convenience, not the
enforcement: unsetting them in the guest only makes the network fail. Leaving
`allowlist` removes them on the next start. The CLI does not write them; set
`http_proxy=http://10.0.2.100:3128` yourself there. The guest settings assume
a Debian-style guest with `sudo`.

What does not work through it:

- `ssh`, `git@github.com:` URLs and `git://`: use `https://` remotes.
- Anything that needs guest DNS or a direct socket: `dig`, `ping`, tools that
  ignore proxy variables.
- Plain-text protocols other than HTTP, and HTTPS to a host not on the list.

What it does not stop: data sent to a listed host. Anything the guest can
write to GitHub (a repository, a gist, an issue) or to a registry it can
publish to leaves the box through the list. The proxy filters on the CONNECT
host and does not read TLS, so it cannot tell which account a request
belongs to, and a listed host that fronts other tenants can be asked for
another one (domain fronting). Keep extra hosts to what the box needs.

## Configuration

Sync is disabled by default. Turn it on per machine with `sync: true`, disable
it again with `false`:

```yaml
sync: true   # opt in: sync the current directory to /workspace on run
```

Without a `sync:` key, or with it `false`, nothing is copied into `/workspace`.
`boite run --no-workspace` overrides an enabled config for a single session.

## Snapshot/rollback

Sandboxes support cheap retries: snapshot the overlay while the VM is stopped,
break things inside the VM, then roll back to the snapshot instead of
recreating the sandbox.

```bash
boite stop my-project
boite snapshot my-project clean
boite start my-project

# ...work, break things, stop again...

boite stop my-project
boite rollback my-project clean
boite start my-project
```

Snapshots are internal qcow2 snapshots on the instance's overlay
(`~/.boite/instances/<name>/overlay.qcow2`), taken with `qemu-img snapshot`.
Create, apply and list all require the VM to be stopped: running qemu holds a
write lock on the overlay that `qemu-img` cannot take. Tags are 1 to 40
characters of letters, digits, dash or underscore.

**Rollback discards the current state.** Everything on the overlay since the
snapshot was taken is gone. The config disk, SSH keys and state file are
untouched, so `boite start` boots the rolled-back disk as usual.

Boite does not resume from a snapshot twice in the same boot session
(QEMU/Firecracker RNG-state caveat): boot the rolled-back disk once, stop the
VM, and roll back again before booting from the same tag another time.

### Sync-out guard

`boite sync` never copies guest-authored hooks or env files into the host
tree. The sync-out deny list skips `.git/hooks/*`, `.envrc`, `.env`, `.env.*`,
`.vscode/tasks.json` and `.husky/*`, plus any regular file carrying the
setuid, setgid or world-writable bit; each skip prints a warning line. The
incoming sync (`boite run`) stays complete. Pass `--all` to disable the guard
for one sync and copy everything.

### Egress

`vm.net` in `~/.boite.yml` picks the guest network. Every mode is enforced
by host processes, so a root process inside the VM cannot undo it.

| Mode | Network | Outbound | Host loopback |
|---|---|---|---|
| `open` (default) | one [passt](https://passt.top) per VM | yes | blocked |
| `offline` | slirp `restrict=on` | none | blocked |
| `allowlist` | slirp `restrict=on` plus a host-side HTTP proxy | listed hosts only, ports 80 and 443 | blocked |
| `open-slirp` | slirp | yes | **reachable**: unsafe |

In `open` mode boite starts `passt --one-off --no-map-gw` for each VM, with
its socket and PID file in the instance directory, and QEMU attaches to it
with `-netdev stream`. `--no-map-gw` stops passt from aliasing the gateway to
the host loopback, so a process in the VM (an agent shell, a prompt-injected
tool) cannot reach services the host binds to `127.0.0.1`: databases,
`boite-worker`, dev servers. Guest DNS still works: passt forwards queries
sent to `10.0.2.3` to the host's first resolver. passt exits when QEMU
disconnects, and `boite stop`, `rm` and a failed start also stop it.

`open` fails with a clear error when `passt` is not installed. boite never
falls back to slirp on its own.

`open-slirp` is the network boite used before passt. QEMU slirp maps the
guest gateway `10.0.2.2` to the host loopback and has no option to turn that
off. Use it only where passt does not exist (macOS, Windows) and nothing
sensitive listens on the host loopback.

Services bound to other host addresses, such as `0.0.0.0` or a Docker
bridge, stay reachable from `open` guests, as they were under slirp. Bind
host services to `127.0.0.1` to hide them from boxes.

`offline` denies the guest all outbound traffic. Inbound SSH through the
host-forwarded port keeps working, so `boite run`, `exec` and `sync` behave
as usual.

```yaml
vm:
  net: offline
```

## Configuration

Boite uses `~/.boite.yml` for configuration. If the file doesn't exist, a default config is created automatically.

```yaml
vm:
  cpus: 2
  memory: 2G
  disk: 20G
  distro: flux             # flux (default) | debian | alpine

sync: true

env:
  source: local            # local | casier
  local:
    vars:                  # literal values baked at create (non-secret config only)
      LOG_LEVEL: debug
    resolve:               # key names pulled by name from the host tiroir store at create
      - OPENROUTER_API_KEY
  casier:
    project: my-org
    environment: dev
    token_ref: CASIER_TOKEN
    ttl: 8h                 # optional: session-scope the materialized token (default: VM lifetime)

provision:
  packages:
    flux:
      - ripgrep
    debian:
      - nala
  commands:
    - 'curl -fsSL https://get.facile.studio | bash'
```

The toolchain, packages, SSH setup and the `boite` user are baked into the base image (see **Default Image**). The config tunes VM resources, sync, and per-instance provisioning. `provision.packages` are installed into a fresh instance with the distro's package manager (`nix profile add` on Flux, apt on Debian, apk on Alpine; a flat list applies to Debian only, keyed lists take `common`, `flux`, `debian` and `alpine`), and `provision.commands` are shell commands run as the `boite` user, both applied once at create right after firstboot. It is the successor to the old `cloud_init.runcmd` / `cloud_init.packages` section: steps run as `boite` over SSH (use `sudo` for anything needing root; the user has passwordless sudo), and the base image already carries the toolchain, so you only add what is not baked in.

### Environment

VMs load an env store sourced by [tiroir](https://github.com/FacileStudio/tiroir) at login. The baked `.zshrc` and `.bashrc` run `eval "$(tiroir export)"`, and `boite exec` re-applies the store for non-interactive commands (which never read an rc file), so managed keys are present in both. Values travel encrypted: tiroir keeps a ciphertext `~/.tiroir` plus a `~/.tiroir.key`, and the store is dropped onto the config disk at create so it exists before the first command — plaintext never sits on the disk or the wire.

- `env.source: local` (default) resolves `env.local.vars` (literal, non-secret) and `env.local.resolve` key names (pulled from the **host** tiroir store) at create time.
- `env.source: casier` (opt-in) materializes a scoped read-only `casier_` project token into the VM's store. Casier is never installed in the guest; the guest holds only a bounded snapshot, never a live token.
- `env.casier.ttl` (optional, e.g. `8h`) makes that token **session-scoped**. With a ttl set, the guest store also holds an expiry marker `<KEY>__expires` (unix-epoch seconds, visible in `boite env list`), and every `boite run` / `boite exec` checks it: an expired token is dropped from the store — stale credentials are never silently kept — and a warning names it once. The next successful refresh re-materializes a fresh token with a new marker; while casier is unreachable the key stays absent and the warning says so. Without `ttl` the token persists for the VM lifetime (the behaviour before session scoping).
- There is **no `sync` command**. `boite run` and `boite exec` re-materialize managed keys before entering the VM, so the store is always current; if casier is unreachable the last snapshot is kept with a warning.
- `boite env set` writes straight into the VM store and **sticks** — a manual value is authoritative over a later source refresh.

Manage the store from the host with `boite env list|get|set|delete <name>`.

## Default Image

Boite boots [Flux](https://github.com/FacileStudio/Flux), the Facile NixOS distribution, in its headless profile: zsh, Neovim, mise, bun, git and the Facile CLIs, with no desktop. `scripts/bake-flux.sh` builds the image with `nix build github:FacileStudio/Flux#boite` (pass a second argument to build another flake ref, such as a local checkout) and falls back to the `nixos/nix` Docker image on hosts without Nix. The image is pinned by URL and SHA256 in `cmd/qemu/distro.go` (`FluxImageSHA256`); rebuild and repin when Flux changes. `nixos` is accepted as an alias of `flux`, so existing configs and instances keep working.

`--distro debian` (or `vm.distro: debian`) selects the baked [Debian 13 (Trixie)](https://www.debian.org/) image built by `scripts/bake-image.sh` via `virt-builder`, which carries the toolchain listed under **Features**. It is pinned in `cmd/qemu/paths.go`. `--distro alpine` selects the Alpine image from `scripts/bake-alpine.sh`.

On every distro the only per-instance input, the SSH public key (or the generated per-VM key), is written onto a small vfat config disk and installed by a firstboot oneshot.

## Features

- **Pre-configured tools** (Debian image): Go 1.26, Bun, Rust, Node 24 with Playwright 1.64.1 and Chromium (`/opt/playwright`, `/opt/ms-playwright`), mise, git, fzf, jq, tmux, wget, make, tiroir, starship
- **Secure VM isolation**: Full virtual machine separation from host
- **Workspace sync**: Current directory copied into /workspace on every `run`
- **Custom shell configuration**: Bakes `.zshrc` and `.tmux.conf` into the base image
- **Interactive zsh shell**: With useful aliases and completions
- **Baked provisioning**: Tool installation and user setup happen at image build time, so `create` is copy-overlay + boot + wait for firstboot

## License

boite is released under the [GNU General Public License v3.0 only](LICENSE) (`GPL-3.0-only`).
