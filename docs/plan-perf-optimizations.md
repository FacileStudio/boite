# Plan — optimize boite's hot paths (perf + code quality)

Status: **done.** Committed 2026-09-11 as 9b70e64 (perf) + d170539 (this
doc). The real-VM e2e pass ran 2026-09-11 against the built v0.7.0 binary:
188MB / 48k-file workspace synced in ~2.1s, byte-identical both directions;
multi-var env refresh materializes every non-pinned key (spaces and resolved
secrets intact), pinned keys stay authoritative, deleted pins re-materialize
on next entry. Gotcha for future sessions: the repo-root `./boite` binary was
an ad-hoc `dev` build predating the env-refresh code, and e2e results taken
from it are wrong — always `go build -o boite .` before a real-VM pass.

Checked against: filet suite gate (CLI pace, no migrations/auth/muse/events —
those suite sections do not apply to a local QEMU CLI). Shell finger-trap:
every remote ssh command must stay a **single shell string** (see
`BuildSSHArgs` + `WithEnv` comment, `cmd/qemu/ssh.go`), and each value must go
through `shellQuote` (`cmd/qemu/provision.go`) so ssh's space-join/reparse and
the remote shell do not tear it apart.

## Goal

Cut the per-command latency of `boite run/create/start` on its hot paths
(loopback workspace sync, env refresh, and the wait loops).

## Why (evidence)

- `cmd/qemu/sync.go:18,22,31,38` gzip over a **loopback** ssh transport — gzip
  is single-threaded CPU that the network never needs; runs on every `sync`.
- `cmd/qemu/env.go:71-75` spawns one ssh process per managed env var on every
  `run` and `exec`.
- `cmd/qemu/helpers.go` `WaitForSSH` adds a fixed 2s; `firstboot.go` polls on a
  fixed 5s cadence.
- `cmd/qemu/image.go` re-hashes the whole ~2.1GB base image on every `create`
  cache hit.

## Approach

Prefer the smallest change that removes wasted work, reusing the existing
single-string ssh contract and `shellQuote`. Raw tar for loopback (zstd noted
but not assumed — the base image has it, arbitrary hosts may not). Batch env
writes into one remote `&&` chain. Drop dead waits; shrivel the wait probes from
fixed cadence to backoff.

## Steps (ordered)

1. `cmd/qemu/sync.go` — raw tar over loopback: `-czf`→`-cf`, `-xzf`→`-xf` in
   `SyncWorkspaceIn`, `SyncWorkspaceOut`. Done.
2. `cmd/qemu/env.go` — `RefreshGuestEnv` batches all keys into one remote
   `tiroir set` chain joined by `&&`, run in a single ssh roundtrip. Done.
3. `cmd/qemu/helpers.go` — remove the hardcoded `time.Sleep(2s)` after
   `WaitForSSH` succeeds. [filet: keep funcLines/returns within limits]. Done.
4. `cmd/qemu/firstboot.go` — `WaitForFirstboot` polls with backoff (~0.5s→2s)
   instead of fixed 5s ticks. Done.
5. `cmd/qemu/sync.go` — stream the tar instead of buffering whole `[]byte`
   (pipe host-tar→ssh / ssh→host-tar), bounded memory, concurrent procs. Done.
6. `cmd/qemu/image.go` — skip full-file SHA-256 on cached base image; verify
   only in `downloadBaseImage` (optionally stat size/mtime on hit). Done.

## Files to Modify

- `cmd/qemu/sync.go` — raw tar (1), streaming (5)
- `cmd/qemu/env.go` — batched env refresh (2)
- `cmd/qemu/helpers.go` — drop fixed 2s (3)
- `cmd/qemu/firstboot.go` — backoff polling (4)
- `cmd/qemu/image.go` — cache-hit checksum skip (6)
- `cmd/qemu/sync_test.go` / `env_test.go` — extend for #5/#2 if a behavior
  worth pinning appears.

## Exit criteria

`go build` clean; `filet check` clean (no raising thresholds); existing tests
pass. Steps 1-2: verify with a real `boite run` that a workspace sync lands
unchanged and a multi-var env refresh still materializes every non-pinned key
(and stops on failure). Step 5: no observable sync regressions on a large
workspace; verified in-tree, real-VM run still TBD before commit/release.
All of the above verified 2026-09-11 (see Status).

## Risks / unknown unknowns

- Raw tar moves more bytes than gz; over loopback to the VM's virtio disk this
  is normally a win, but for a very large, highly-compressible tree gz could
  win. If so, revisit with `-I zstd -T0` (base image ships zstd).
- Batching env with `&&` stops at the first failing key, same as today's
  per-key early return — the offline-fallback contract (caller warns, keeps
  last snapshot) is preserved. Error message loses the failing key name; keep
  it generic.

## Skip (YAGNI)

- No `-I zstd` default (host presence not guaranteed on darwin/windows).
- No async/proc-pool redesign of `WaitForPID`/`IsProcessRunning`; per-file
  `cmd/shutdown` etc. are out of scope. Shrink only the confirmed fixed costs.