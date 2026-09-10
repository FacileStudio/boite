# Summary of Fixes for boite VM Configuration

> Superseded: the `cloud_init:` config, `MergeCloudInitConfig`, `GenerateSeedISO`
> and the cloud-init seed ISO described below were removed in the baked-image
> refactor. Provisioning now lives in the base image built by
> `scripts/bake-image.sh`; the only per-instance input is the SSH public key,
> delivered on a config ISO and installed by a firstboot oneshot. Read `README.md`
> for the current model. The notes below are kept as history.

## Issues Fixed

1. **[config]** `--config` flag now properly honored
   - Fixed `cmd/root.go` and `cmd/qemu/config.go` so explicit config paths are passed through
   - `LoadBoiteConfig()` now accepts a path parameter and falls back to `~/.boite.yml` when empty
   - Config path is threaded through `Create()`, `Start()`, and `GenerateSeedISO()` functions

2. **[arch]** VM resources (`cpus`, `memory`, `disk`) now applied from config
   - Added `parseDiskGB()` helper to handle disk size parsing (e.g., "5G" -> 5)
   - VM config values override hardcoded defaults in both `Create()` and `Start()` paths
   - Disk size for overlay creation now respects `vm.disk` setting

3. **[arch]** `write_files`, `packages`, and `runcmd` apply at VM creation time
   - These sections from `~/.boite.yml` (or explicit `--config`) are merged into cloud-init seed ISO
   - Uses existing `MergeCloudInitConfig()` logic to combine user config with defaults
   - Applied during `boite create` when the seed.iso is generated

4. **[bug]** Fixed panic in `StartQEMU` on console log creation failure
   - Moved `defer f.Close()` inside success path in `cmd/qemu/qemu.go`
   - Prevents nil pointer dereference when `os.Create()` fails

5. **[behavior]** SSH keys now instance-scoped (already implemented in prior diff)
   - Keys stored in instance directory rather than shared `~/.boite/ssh/`
   - Each VM gets unique SSH key pair for better isolation

## Verification

- All Go code compiles: `go build ./...`
- All tests pass: `go test ./...`
- Config plumbing verified: `--config /path/to/config.yml create test-vm --no-mount` attempts to read the specified file
- No regressions in existing functionality

## Files Modified

- `cmd/root.go` - config flag handling and initialization
- `cmd/create.go` - pass config path to qemu.Create
- `cmd/start.go` - pass config path to qemu.Start
- `cmd/qemu/config.go` - updated LoadBoiteConfig to accept path parameter
- `cmd/qemu/cloudinit.go` - pass config path to GenerateSeedISO and LoadBoiteConfig
- `cmd/qemu/lifecycle.go` - 
  - Added config path parameters to Create() and Start()
  - Implemented VM resource parsing (cpus, memory, disk)
  - Fixed SSH key passphrase handling from VM config
  - Added parseDiskGB() helper
- `cmd/qemu/qemu.go` - fixed panic in StartQEMU by moving defer inside success path

## Behavior Changes

**Before**: 
- `--config` flag was parsed but ignored
- VM resources always defaulted to 2 CPUs, 4G memory, 10G disk
- `write_files`, `packages`, `runcmd` only worked from `~/.boite.yml` (not explicit config)
- Panic possible if console log creation failed

**After**:
- `--config /path/to/file.yml` is respected for all configuration sections
- VM resources (cpu, memory, disk) read from config file
- All cloud-init sections (write_files, packages, runcmd, users, apt) apply from specified config
- No panic on console log creation failure
- Instance-scoped SSH keys remain (improved security isolation)

The fix ensures that `boite create --config /path/to/config.yml myvm` will:
1. Read VM settings (cpus, memory, disk, ssh_key_passphrase) from the config
2. Apply write_files, packages, and runcmd from the config to the seed ISO
3. Respect any custom APT sources or user definitions in the config
4. Use the specified disk size for the VM overlay
5. Start the VM with configured resources