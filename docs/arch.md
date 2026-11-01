# Architecture decisions

## The worker is a host daemon, separate from the web tier

Status: accepted. Date: 2026-11-01.

### Context

A web dashboard (boite-web, Bulle) needs to create, start, stop and snapshot QEMU boxes for signed-in users. The CLI already owns that lifecycle in `cmd/qemu`. Two shapes were possible: run QEMU inside the web process, or run it in a separate host process the web tier calls.

### Decision

`boite-worker` is a systemd service on the host. It runs as root, owns `/dev/kvm` and the box pool in `/var/lib/boite`, and exposes a token-authenticated HTTP API on a Unix socket (mode 0600). The web tier runs in its normal minimal container and mounts that socket.

### Why

- Boxes survive web redeploys. QEMU processes are children of the host service, not of the container, so `docker compose up -d --build` restarts the dashboard and leaves every box running. The unit uses `KillMode=process` so even a worker restart keeps them.
- A web-tier breach does not reach the hypervisor. The container has no `/dev/kvm`, no QEMU and no `--privileged`. It can only ask the worker for the operations in its API, and only with the token.
- One lifecycle implementation. The worker imports `cmd/qemu` and reuses `Create`, `Start`, `Stop`, `Destroy`, the snapshot code, `FindFreePort` and `state.json`. The CLI and the worker cannot drift.

### Consequences

- Two deploy units. The web tier ships with `docker compose`; the worker is installed and upgraded by hand (see the README worker section).
- Box state lives on one host's disk. Overlays are not portable; backup and move are out of scope.
- SSH ports are host ports. `BOITE_SSH_BIND` picks the interface; bind a private address, never a public wildcard.
- Idle policy stays in the caller. The worker only offers `stop`.
- The control key (`BOITE_CONTROL_PUBKEY`) is written into every box so the caller can reach it; its private half stays with the caller.
- Creates are asynchronous. `POST /boxes` persists the box as `creating` and returns 202; a single queue runs creates one at a time, so two boxes never pick the same SSH port. A restart marks unfinished creates as `error`.
