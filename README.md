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

QEMU, `mkfs.fat` and `mcopy` are required for VM-based sandboxes.

- Ubuntu/Debian: `sudo apt install qemu-kvm dosfstools mtools`
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
| `install` | Install boite to system PATH |

## Workspace sync

Workspace sync is an opt-in feature. When enabled, `boite run` copies the
directory you run it from into the sandbox's `/workspace` before opening the
shell, so the VM sees the code you are working on. The copy overwrites
`/workspace` wholesale (any changes made inside the VM are replaced), so push
changes back with `boite sync` before re-running.

The sync is a plain file copy over SSH, not a mount: the VM's `/workspace` is a
bounded snapshot, never a live handle into your host tree. This is what keeps a
compromised guest from reaching your files.

### Configuration

Sync is disabled by default. Turn it on per machine with `sync: true`, disable
it again with `false`:

```yaml
sync: true   # opt in: sync the current directory to /workspace on run
```

Without a `sync:` key, or with it `false`, nothing is copied into `/workspace`.
`boite run --no-workspace` overrides an enabled config for a single session.

## Configuration

Boite uses `~/.boite.yml` for configuration. If the file doesn't exist, a default config is created automatically.

```yaml
vm:
  cpus: 2
  memory: 2G
  disk: 20G

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

provision:
  packages:
    - nala
  commands:
    - 'curl -fsSL https://get.facile.studio | bash'
```

The toolchain, packages, dotfiles, SSH host keys and the `boite` user are baked into the base image by `scripts/bake-image.sh` (see **Default Image**). The config tunes VM resources, sync, and per-instance provisioning. `provision.packages` are apt packages installed into a fresh instance, and `provision.commands` are shell commands run as the `boite` user, both applied once at create right after firstboot. It is the successor to the old `cloud_init.runcmd` / `cloud_init.packages` section: steps run as `boite` over SSH (use `sudo` for anything needing root; the user has passwordless sudo), and the base image already carries the toolchain, so you only add what is not baked in.

### Environment

VMs load an env store sourced by [tiroir](https://github.com/FacileStudio/tiroir) at login. The baked `.zshrc` and `.bashrc` run `eval "$(tiroir export)"`, and `boite exec` re-applies the store for non-interactive commands (which never read an rc file), so managed keys are present in both. Values travel encrypted: tiroir keeps a ciphertext `~/.tiroir` plus a `~/.tiroir.key`, and the store is dropped onto the config disk at create so it exists before the first command — plaintext never sits on the disk or the wire.

- `env.source: local` (default) resolves `env.local.vars` (literal, non-secret) and `env.local.resolve` key names (pulled from the **host** tiroir store) at create time.
- `env.source: casier` (opt-in) materializes a scoped read-only `casier_` project token into the VM's store. Casier is never installed in the guest; the guest holds only a bounded snapshot, never a live token.
- There is **no `sync` command**. `boite run` and `boite exec` re-materialize managed keys before entering the VM, so the store is always current; if casier is unreachable the last snapshot is kept with a warning.
- `boite env set` writes straight into the VM store and **sticks** — a manual value is authoritative over a later source refresh.

Manage the store from the host with `boite env list|get|set|delete <name>`.

## Default Image

Boite uses a baked [Debian 13 (Trixie)](https://www.debian.org/) image built by `scripts/bake-image.sh` via `virt-builder`. The bake installs the toolchain, creates the `boite` user, regenerates SSH host keys, configures the network and lays down the dotfiles, then purges cloud-init. At runtime the only per-instance input — the SSH public key (or the generated per-VM key) — is written onto a small vfat config disk and installed by a firstboot oneshot. The baked image is pinned by URL and SHA256 in `cmd/qemu/paths.go`; rebuild and repin when Debian or the toolchain changes.

## Features

- **Pre-configured tools**: Go 1.26, Bun, Rust, mise, git, fzf, jq, tmux, wget, make, tiroir, starship
- **Secure VM isolation**: Full virtual machine separation from host
- **Workspace sync**: Current directory copied into /workspace on every `run`
- **Custom shell configuration**: Bakes `.zshrc` and `.tmux.conf` into the base image
- **Interactive zsh shell**: With useful aliases and completions
- **Baked provisioning**: Tool installation and user setup happen at image build time, so `create` is copy-overlay + boot + wait for firstboot

## License

MIT
