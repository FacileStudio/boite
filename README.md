# Boite - Development Sandbox Manager

CLI tool to create, manage, and clean up virtual machine-based development sandboxes for general development.

## Install

### Option 1: Build and use locally

```bash
# Navigate to the boite directory
cd Code/Facile/boite

# Build the CLI
go build -o boite .

# Use directly
./boite
```

### Option 2: Install globally (recommended)

```bash
# Navigate to the boite directory
cd Code/Facile/boite

# Build the CLI
go build -o boite .

# Install to your local bin or system PATH
./boite install

# Or with sudo for system-wide installation
sudo ./boite install
```

After installation, you can run `boite` from anywhere:

```bash
boite
```

## Prerequisites

**For VM-based sandboxes (recommended):**
```bash
# Ubuntu
sudo snap install multipass

# macOS
brew install --cask multipass

# Windows
winget install Canonical.Multipass
```

**For Docker-based sandboxes (alternative):**
- Docker daemon running
- User in docker group (`sudo usermod -aG docker $USER`)

## Usage

### Create and manage sandboxes

```bash
# List existing sandboxes
boite list

# Create a new sandbox VM
boite create my-project

# Access a sandbox VM
boite exec my-project

# Stop a sandbox (preserves data)
boite stop my-project

# Remove a sandbox permanently
boite rm my-project
```

## Features

- **Pre-configured tools**: Go 1.26, Bun, Rust, mise, git, fzf, jq, tmux, wget, make, skatos
- **Secure VM isolation**: Full virtual machine separation from host
- **User-mapped VMs**: Uses your username, UID, and GID by default
- **Workspace mounting**: Current directory mounted at /workspace
- **Interactive zsh shell**: With useful aliases and completions
- **tmux support**: Built-in terminal multiplexer configuration
- **Automatic provisioning**: Cloud-init handles all tool installation

## Commands

| Command | Description |
|---------|-------------|
| `build [tag]` | Pre-build a VM image using Multipass |
| `create [name]` | Create a new sandbox VM |
| `list` | List all sandboxes |
| `exec [name]` | Enter sandbox VM shell |
| `stop [name]` | Stop sandbox VM (preserves data) |
| `rm [name]` | Remove sandbox VM permanently |
| `install` | Install boite to system PATH |

## VM Image

Boite creates Ubuntu 24.04 VMs with automatic provisioning:

```bash
boite build dev-sandbox
```

The VM is pre-configured with:
- **mise** for version management
- **Go 1.26**
- **Rust** (stable via rustup)
- **Bun 1.3.14**
- **Skatos** for secrets management
- **tmux** for terminal multiplexing
- **fzf** for fuzzy finding
- **jq** for JSON processing
- **Oh-my-zsh** with basic configuration
- **Non-root user** (UID/GID mapped from host)

## Security

- Full VM isolation (not container sharing kernel)
- Runs as non-root user inside VM
- Minimal attack surface with essential tools only
- Workspace is the only mounted directory by default

## Requirements

- Go 1.26 (to build the CLI)
- Multipass (for VM-based sandboxes) OR Docker (for container-based)

## Integration with Facile CLI

This tool integrates with the Facile CLI suite:

```bash
# Once facile is installed
facile install boite
```

## License

MIT
