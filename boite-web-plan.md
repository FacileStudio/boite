# Boite Web: self-hosted dev-box dashboard

Implémentation plan (cold-start handoff — a fresh worker needs no prior conversation to execute this).

Status: APPROVED — architecture decision locked: **Option A, host daemon worker + separate web tier**. The decision rationale (why A over the single-process B) is in `docs/arch.md`. Not yet executed.

## Goal

A self-hostable Facile suite web app that manages online development VMs: you log in via porte SSO, create a QEMU dev box from a dashboard, paste your SSH public key, and get back `<host>:<port>` + `user:boite` to connect to with your own key. The whole thing deploys on one host alongside the other suite apps.

## Why (evidence) — two lines

`boite` (this repo) already has the two hard parts: a baked base image + config-ISO SSH-key injection (`scripts/bake-image.sh`, `cmd/qemu/firstboot.go`) that boots a provisioned VM in seconds, and a full QEMU lifecycle (`cmd/qemu/lifecycle.go`). What is missing is any way to drive that from a browser, authenticated, behind the suite's own auth/log/env/deploy stack. The CLI's SSH is loopback-only (`hostfwd=tcp:127.0.0.1:<port>-:22`, ports scanned 2222–2322 in `cmd/qemu/instance.go`), so it is not remotely reachable as-is. And a web redeploy must never take running dev VMs down with it — that is the reason the VM engine is a separate host process, not part of the web tier.

## Approach

Two processes. The host one that owns QEMU is deliberately not the one users' web requests hit.

- **boite-web** — the control plane, a new suite repo (`~/Code/Facile/boite-web`, module `github.com/FacileStudio/boite-web/apps/api`), TS-family layout: `apps/api` (Go) + `apps/client` (SvelteKit 5, muse). It is the Sablier shape: tronc chassis (`/health`, `/ready`, SPA serving, env), porte OIDC SSO + local login + device grant, Journal SDK logging + browser error reporting, casier env, real PostgreSQL. It stores box metadata and user keys in PG and calls the worker over a control channel. It does **not** touch `/dev/kvm` — it cannot see QEMU.
- **boite-worker** — a host daemon (systemd, root, outside any container). It owns `/dev/kvm` and the `/var/lib/boite/` instance pool, reuses boite's `cmd/qemu/*` as a library, and exposes a tiny token-authenticated HTTP control API on a Unix socket: `create / start / stop / rm / list / set-ssh-key / connect-info`.

The API calls the worker over the Unix socket (mounted into the container). Only the API reaches the worker; a box's SSH endpoint is a host `:port` that QEMU forwards, and a customer SSH key never touches boite's own control SSH.

Why separate: the suite's web tier stays in its normal minimal distroless image (no QEMU, no `--privileged`), so a web-tier breach does not reach the hypervisor trust domain; and `docker compose up -d api` redeploys the dashboard without restarting the worker, so **live boxes survive a web redeploy**.

## Steps (ordered)

### Track A — worker (host engine), depends on nothing
1. `boite/cmd/worker.go` (new binary in this repo, imports `cmd/qemu/*`) — a small HTTP control server bound to a Unix socket (path from env `BOITE_WORKER_SOCK`, default `/run/boite/worker.sock`), authorized by a bearer token from env `BOITE_WORKER_TOKEN`. Endpoints map 1:1 onto the existing `lifecycle.go` functions: `POST /boxes` (create), `POST /boxes/:id/start|stop|rm`, `GET /boxes`, `GET /boxes/:id` (connect info: ssh port, user, host), `PUT /boxes/:id/ssh-key` (rebuild config ISO via `BuildConfigISO`, restart box). Reuse `FindFreePort`/state.json verbatim; do not fork them. A small start-block runs the token/socket setup and post-start QEMU calls. [exit: `go build ./...` clean; run against a temp `BOITE_INSTANCE_DIR`, `curl --unix-socket` boots a box and returns connect info]
2. `boite/cmd/qemu/instance.go` — make the SSH bind address configurable (default `127.0.0.1`; the daemon sets `BOITE_SSH_BIND` to `0.0.0.0` or an explicit host address so boxes are reachable remotely). Keep the 2222–2322 scan. [exit: one line threads config into `hostfwd`; loopback default unchanged for the CLI]
3. `scripts/boite-worker.service` (new) — systemd unit: `ExecStart=/usr/local/bin/boite-worker`, `User=root`, socket directory + `StateDirectory=boite`, `Restart=on-failure`, after-network. Document the host prereq `apt install qemu-kvm dosfstools mtools genisoimage` (mirrors `README.md`). [exit: `systemctl enable --now boite-worker` brings the socket up]

### Track B — control plane (new repo `boite-web`), depends on Track A step 1 for the socket contract
4. Scaffold `~/Code/Facile/boite-web` from a current suite reference (`Sablier` layout): `apps/api` (Go, chi + tronc + porte + GORM + journal sdk), `apps/client` (SvelteKit 5 + muse), root `Dockerfile` (distroless single image), `docker-compose.yml`, `casier.yml` (`project.slug: boite-web`, `environment: dev`), `.env.example`, `roofilet.yml`. Module `github.com/FacileStudio/boite-web/apps/api`. [exit: `make build` clean, `/health` answers, client builds]
5. `apps/api/schemas/` — box + key models: `boxes` (id, name, status, vm_cpus/vm_memory/vm_disk, ssh_local_port, created_at, started_at), `box_keys` (id, box_id, public_key, fingerprint, added_at). GORM models + migrations, auto-run at startup (`schemas.Migrate(db)`, tronc/migrate). [exit: migrations run clean against a fresh PG; `goose_db_version` not hand-created]
6. `apps/api/main.go` — tronc assembly + porte kit (mirror Sablier `buildAuth`): shared session manager, OIDC SSO kit + local login, `SSO_ONLY`, device grant (`OIDC_CLI_AUDIENCE`) so `facile login` works, Journal SDK init (`JOURNAL_URL`/`JOURNAL_TOKEN` server, `JOURNAL_BROWSER_URL`/`JOURNAL_BROWSER_KEY` browser), casier env loading. [exit: SSO login works, requests land structured in Journal, `/health` green]
7. `apps/api/modules/boxes/` — controller/router/service/types + `porte.go` (porte auth floor: cookie before Authorization, `X-Facile-CSRF` on cookie mutations). Routes: `GET /api/boxes`, `POST /api/boxes` (create), `GET /api/boxes/:id` (connect info = host + ssh_local_port + user `boite`), `PUT /api/boxes/:id/ssh-key`, `POST /api/boxes/:id/start|stop|rm`. Handlers proxy to the worker socket and persist the row + port in PG. [exit: curl through porte succeeds; box row + returned port match the worker]
8. `apps/api/internal/worker/` (new) — typed worker client: reads socket path + token from casier env, JSON-over-socket calls, maps worker errors to the suite error envelope, idempotent self-correction (if PG says `running` but the worker does not, reconcile). [exit: unit tests with a stubbed worker socket pass]
9. `apps/client/` (muse) — `login/` (porte), `(app)/` layout with `dashboard/` (list + create box, paste SSH key, show `<host>:<port>` + user `boite` + copyable `ssh -p <port> boite@<host>` command) and `settings/` (default vm resources, ssh bind host). `$lib/backend.ts` fetch wrapper; muse tokens, Svelte 5 runes, no legacy stores. [exit: `filet` clean; dashboard create→connect flow works end to end]
10. `docker-compose.yml` + `Dockerfile` — db (postgres 16) + api (distroless, serves client via tronc `spa`), dokploy/traefik labels, `TRUSTED_PROXIES`. Mount the worker socket into the api service (`/run/boite/worker.sock`) and set the worker token. The **worker is deliberately not in compose** — it is the host systemd unit from Track A. [exit: `docker compose up` on a host with a running worker reaches the dashboard behind traefik]

### Common gate
11. `sh scripts/check.sh` (gofmt, `go vet`, `go test`, `filet check .`) green in both repos; one manual end-to-end: create box → paste key → `ssh` from a laptop through the dashboard's published host:port; then `docker compose up -d --build api` and confirm a running box stays up. [exit: check.sh clean, real SSH connect with a user-supplied key, box survives a web redeploy]

## Files to Modify / New

- `boite/`: `cmd/worker.go` (new), `cmd/qemu/instance.go` (bind address), `scripts/boite-worker.service` (new), `README.md` (self-host + worker section), `docs/arch.md` (worker/web separation decision)
- `boite-web/` (new repo): `apps/api/main.go`, `apps/api/modules/boxes/*`, `apps/api/internal/worker/*`, `apps/api/schemas/*`, `apps/client/src/routes/login/*`, `apps/client/src/routes/(app)/dashboard/*`, `apps/client/src/routes/(app)/settings/*`, `apps/client/src/lib/backend.ts`, `Dockerfile`, `docker-compose.yml`, `casier.yml`, `.env.example`, `roofilet.yml`

## Exit criteria

- `boite-web` is a deployable self-hosted app: SSO login (porte), boxes in a dashboard, create → get `<host>:<port>` + `user:boite` → paste your own SSH key → connect from a laptop.
- Requests appear as structured logs in the suite Journal; `facile login` device flow works.
- The worker socket is reachable only by the API container; customer keys never reach boite's own control SSH.
- A web redeploy (`docker compose up -d --build api`) does not kill running boxes.
- `filet` and `scripts/check.sh` green in both repos.

## Risks / unknown unknowns

- **Two deploy units to manage.** The web half is `docker compose`, but the worker is a host systemd unit that dokploy does not manage — install/upgrade it by hand (the unit's prereq line covers it). This is the accepted price of keeping boxes alive across web deploys.
- **Socket token secret.** Only the API container holds `BOITE_WORKER_TOKEN` (via casier); the worker binds the socket `0600`. Orphan the token on rotation; a disposing container should not leave boxes behind — gate on `BOITE_WORKER_TOKEN` change.
- **Remote SSH reachability.** slirp hostfwd gives each box a **port on the VM host**, not a per-VM IP. Bumping `BOITE_SSH_BIND` from loopback to the host LAN/public interface is the whole change, but it exposes a range of raw ports — defaulting to the host LAN interface, not world-open. Whether a bastion hop is needed per deployment is a deployment decision, not a v1 gate.
- **How user keys are applied to already-running boxes.** `BuildConfigISO` + firstboot oneshot runs at create. Applying a key to a running box means a stop/restart or a second ISO. Confirm the UX (apply-on-next-restart vs live) when wiring step 7.
- **Box state lives on one host's disk.** Copy-on-write overlays are cheap but not portable. Backups/move are out of scope for v1 but the daemon's disk layout (`StateDirectory=boite`) must not block them.
- **Cross-repo dependency.** `boite-web` depends on the boite worker surface; publish via `github:FacileStudio/boite#<branch>` / a versioned tag when the contract stabilizes, per `[distribute]`.

## Skip (YAGNI)

- **Saas concerns for v1**: billing, quotas, multi-tenant isolation, per-box public IPs, idle cost metering. The product is self-hosted.
- **Per-box public IPs** — a `host:port` + user is enough; IP-per-box is expensive and unneeded.
- **Bastion/gateway host** — add only if raw port exposure is unacceptable in a target deployment.
- **Live key application to running boxes** — a restart-based application (or create-time-only) is v1; live reload is a follow-on.
- **Spaces/multi-tenancy** (`porte/spaces`) — single-owner host for v1; each self-hoster runs their own instance.
- **Box persistence/migration tooling** — backup/restore across hosts is a later feature, not a v1 gate.

## Convention flags checked

Checked against: `[migrations]` (GORM `schemas.Migrate`, tronc/migrate, no hand-created `goose_db_version`), `[auth/porte]` (cookie before Authorization, `X-Facile-CSRF`, `SSO_ONLY`, device grant, `email_verified:false` trap), `[muse]` (muse tokens, Svelte 5 runes, no legacy stores), `[module-path]` (`github.com/FacileStudio/boite-web/apps/api`), `[filet]` (gate clean, don't raise thresholds), `[distribute]` (cross-repo via `github:FacileStudio/boite#branch`), `[events]` (the auth/audit events use the `@facile/events` envelope keyed on `actor_email`).

Not applicable: `[spaces]` (single-owner v1), `[casier]` applies (worker token/socket path live in `casier.yml` env).