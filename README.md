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

QEMU with libvirt and cloud-init are required for VM-based sandboxes.

- Ubuntu/Debian: `sudo apt install qemu-kvm libvirt-daemon-system cloud-init`
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

Sync is disabled by default. Turn it on per machine with `workspace.sync_at_run:
true`, disable it again with `false`:

```yaml
workspace:
  sync_at_run: true   # opt in: sync the current directory to /workspace on run
```

Without a `workspace:` section, or with `sync_at_run` absent or `false`, nothing
is copied into `/workspace`. `boite run --no-workspace` overrides an enabled
config for a single session.

## Configuration

Boite uses `~/.boite.yml` for configuration. If the file doesn't exist, a default config is created automatically.

```yaml
vm:
  cpus: 2
  memory: 2G
  disk: 20G

cloud_init:
  packages:
    - git
    - curl
    - wget
    - unzip
    - ca-certificates
    - bash-completion
    - fzf
    - jq
    - make
    - build-essential
    - zsh
    - neovim
    - starship
    - tmux
    - htop
    - docker.io

  write_files:
    - path: /home/boite/.zshrc
      permissions: "0644"
      owner: boite:boite
      content: |
        #!/bin/zsh
        # Boite workspace shell config

        # Paths
        export PATH="/usr/local/go/bin:$HOME/.cargo/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.bun/bin:$HOME/.local/share/mise/shims:$PATH"
        export WORKSPACE=/workspace

        # Mise activation
        if command -v mise >/dev/null 2>&1; then
          eval "$(mise activate zsh)"
        elif [ -x /usr/local/bin/mise ]; then
          eval "$(/usr/local/bin/mise activate zsh)"
        fi

        # Starship prompt
        if command -v starship >/dev/null 2>&1; then
          eval "$(starship init zsh)"
        fi

  runcmd:
    - rm -f /etc/legal /etc/motd
    - touch /home/boite/.hushlogin
    - curl https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh
    - curl -L https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -
    - su - boite -c 'curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y'
    - su - boite -c 'curl -fsSL https://bun.sh/install | bash'
    - su - boite -c 'curl -fsSL https://raw.githubusercontent.com/saravenpi/skatos/main/install.sh | bash'
    - su - boite -c 'git config --global init.defaultBranch main'
    - su - boite -c 'git config --global safe.directory "*"'
    - mkdir -p /workspace
    - chown -R boite:boite /workspace
```

Your `cloud_init:` config is merged with the repo's built-in provisioning in `cmd/qemu/cloudinit.yml` (mise, Go, Rust, bun, skatos, plus `.zshrc` and `.tmux.conf`). There is no top-level `users:` key; it is silently ignored. Define users under `cloud_init.users`.

## Default Image

Boite uses [Debian 13 (Trixie) cloud image](https://cloud.debian.org/images/cloud/trixie/latest/) as the base for new VMs.

## Features

- **Pre-configured tools**: Go 1.26, Bun, Rust, mise, git, fzf, jq, tmux, wget, make, skatos, starship
- **Secure VM isolation**: Full virtual machine separation from host
- **Workspace sync**: Current directory copied into /workspace on every `run`
- **Custom shell configuration**: Provisions `.zshrc` from config
- **Interactive zsh shell**: With useful aliases and completions
- **Automatic provisioning**: Cloud-init handles tool installation

## License

MIT
