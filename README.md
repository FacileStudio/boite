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

Multipass (https://multipass.run/):
- Ubuntu: `sudo snap install multipass`
- macOS: `brew install --cask multipass`
- Windows: `winget install Canonical.Multipass`

## Usage

```sh
# Create a new sandbox VM with current directory mounted at /workspace
boite create my-project

# Open an interactive shell
boite shell my-project

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

# Purge all deleted sandboxes
boite purge
```

## Commands

| Command | Description |
|---------|-------------|
| `create <name>` | Create a new sandbox VM |
| `shell [name]` | Open interactive shell in sandbox VM |
| `exec [name] [cmd...]` | Execute command or shell in sandbox VM |
| `list` | List all sandboxes |
| `start <name>` | Start sandbox VM |
| `stop [name]` | Stop sandbox VM (preserves data) |
| `remove [name]` | Remove sandbox VM permanently |
| `purge` | Purge all deleted sandbox VMs |
| `install` | Install boite to system PATH |

## Features

- **Pre-configured tools**: Go 1.26, Bun, Rust, mise, git, fzf, jq, tmux, wget, make, skatos
- **Secure VM isolation**: Full virtual machine separation from host
- **Workspace mounting**: Current directory mounted at /workspace
- **Custom shell configuration**: Provisions `.zshrc_local` from workspace if present
- **Interactive zsh shell**: With useful aliases and completions
- **Automatic provisioning**: Cloud-init handles tool installation

## License

MIT
