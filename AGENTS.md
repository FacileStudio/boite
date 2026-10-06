# AGENTS.md - Boite Project

## Overview

**Boite** is a development sandbox manager – a CLI tool that creates, manages, and cleans up QEMU-based virtual machine development environments. It's designed as a self-hosted solution for developers who need isolated dev VMs with secure environment variable management.

### Core Components

| Component | Purpose |
|-----------||----------|
| `boite` (CLI) | Main tool – `boite create`, `boite run`, `boite exec`, `boite list`, `boite stop`, `boite start`, `boite rm`, `boite snapshot`, `boite rollback`, `boite sync` |
| `boite-web` (new repo) | Self-hosted web dashboard (SvelteKit + Muse) for managing VMs, creating instances, and viewing status |
| `tiroir` (new repo) | Encrypted environment variable store – loads secrets into VMs at creation time, supports `boite env` CRUD commands |
| `boite/cmd/worker` | Host-side QEMU daemon (systemd service) that provides the control plane API |
| `boite/cmd/qemu/` | QEMU lifecycle management – instance creation, SSH binding, snapshot/rollback |

### Architecture

- **Host side** – `boite` CLI + `boite-worker` (systemd daemon) – owns QEMU, manages VM lifecycle, exposes a Unix socket API
- **Web side** – `boite-web` (API + SPA) – dashboard UI for creating/starting/listing VMs, applying SSH keys
- **Env side** – `tiroir` – encrypted env store (ciphertext + key) that injects secrets into VMs at creation time

### Quick Start

```bash
# Install (via Facile or from source)
facile install boite

# Create a new sandbox VM
boite create my-project

# Open an interactive shell (syncs current dir into /workspace)
boite run my-project

# List all sandboxes
boite list

# Stop a sandbox (preserves data)
boite stop my-project

# Start a stopped sandbox
boite start my-project

# Execute a command inside the sandbox
boite exec my-project ls -la

# Remove a sandbox permanently
boite rm my-project
```

### Key Commands

| Command | Description |
|---------|-------------|
| `boite create <name>` | Create a new VM with default resources (2 CPU, 2GB RAM, 20GB disk) |
| `boite run <name>` | Launch the VM and open an interactive shell (syncs workspace) |
| `boite list` | List all managed sandboxes |
| `boite stop <name>` | Gracefully stop a running VM |
| `boite start <name>` | Start a previously stopped VM |
| `boite exec <name> [cmd...]` | Execute a command inside the running VM |
| `boite snapshot <name> <tag>` | Take a checkpoint of a stopped VM |
| `boite rollback <name> <tag>` | Restore a VM to a previous snapshot |
| `boite rm <name>` | Permanently delete a VM (destroys disk) |
| `boite env <name> <key>` | Set an environment variable in a VM |
| `boite env <name> list` | List all env vars for a VM |
| `boite env set <name> <key> <value>` | Set an env var (authoritative over source refresh) |
| `boite env get <name> <key>` | Get an env var from a VM |

### Environment Management (tiroir)

The `tiroir` project provides a secure, encrypted environment variable store:

- **Local source** – Secrets baked into the VM at creation (via `boite env set`)
- **Casier source** – Optional scoped token from the `casier` env system (opt-in)
- **Manual override** – `boite env set` writes directly into the guest store and sticks across re-entries

#### Example Workflow

```bash
# Set a secret for a VM
boite env my-project DB_PASSWORD secret_value

# Verify it's present
boite env my-project DB_PASSWORD

# View all env vars
boite env my-project list
```

### Configuration

Configuration is stored in `~/.boite.yml` (or `~/.boite.yml` in the project root):

```yaml
vm:
  cpus: 2
  memory: 2G
  disk: 20G

sync: true  # Opt-in: copy current dir into /workspace on run

env:
  source: local          # local | casier
  local:
    vars:
      LOG_LEVEL: debug
  casier:
    project: my-org
    environment: dev
    token_ref: CASIER_TOKEN
```

### Security Notes

- **Secrets never leave the host** – `tiroir` encrypts at rest (ciphertext + random key) and never transmits plaintext
- **SSH keys are injected securely** – `boite exec` uses `WithEnv` to paint `tiroir export` into the shell
- **No `sync` command** – env is refreshed automatically at the host→guest boundary; `boite run`/`boite exec` re-materialize managed keys
- **Worker is a host systemd unit** – not managed by `docker compose`; ensure `boite-worker` is running for the socket to be available

### Troubleshooting

| Issue | Cause | Fix |
|-------|-------|-----|
| `boite run` fails with “QEMU not found” | QEMU not installed or not in PATH | Install `qemu-system-x86_64 qemu-img mcopy` |
| `boite env set` silently fails | Missing `tiroir` binary or permissions | Ensure `tiroir` is in `PATH` and owned by `boite:boite` |
| VM won't start after `boite create` | Network bind address mismatch | Check `BOITE_SSH_BIND` env var; default is `127.0.0.1` |
| `boite exec` doesn't see env vars | Guest RC not sourced | `boite exec` wraps the command with `eval "$(tiroir export)"` |
| `boite list` shows no VMs | No VMs created or not yet started | Run `boite create <name>` first |

### Related Projects

- **boite-web** – Dashboard for managing VMs (new repo, `github.com/FacileStudio/boite-web`)
- **tiroir** – Encrypted env store (new repo, `github.com/FacileStudio/tiroir`)
- **facile** – Base framework (Go 1.26, bun 1.3.14, rust managed by system)

### References

- `README.md` – Full usage documentation
- `boite/cmd/worker.go` – Host QEMU daemon
- `boite/cmd/qemu/lifecycle.go` – Instance lifecycle (create, start, stop, snapshot)
- `tiroir/cmd/tiroir/main.go` – Env CRUD CLI
- `scripts/check.sh` – Code quality gate (gofmt, vet, test, filet)
- `boite-web/` – Web dashboard (SvelteKit + muse)

### Version History

- **v0.5.0** (2026-09-10) – Tiroir env system shipped; boite-web and boite-web-deploy planned
- **v0.4.x** – Core boite CLI with VM lifecycle
- **v0.3.x** – Initial QEMU integration and workspace sync
- **v0.2.x** – Tiroir env system beta

### Running Tests

```bash
make test          # go test ./...
make check         # scripts/check.sh (gofmt, vet, test, filet)
```

### Memory

This project is tracked in `~/.mycelium/memory/` with the following structure:
- `overview.md` – Core memory summary
- `index.md` – Router to all pages
- `log.md` – Append-only operation log
- `bugs/`, `tools/`, `projects/`, `conventions/`, `standards/`, `syntax/` – Categorized knowledge

## Getting Help

- **Issue**: Describe the problem clearly, include error output if any
- **Debug**: Run `boite env list` to see current env state, `boite list` to see VMs
- **Security**: All secrets are handled by `tiroir` – never printed in logs
- **Support**: Contact the boite maintainers via the project's issue tracker

## License

MIT
