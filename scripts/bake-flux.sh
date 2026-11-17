#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

out="${1:-flux.qcow2}"
flake="${2:-github:FacileStudio/Flux}"
attr="packages.x86_64-linux.boite"
bake="flux.baked.qcow2"
nix_args=(--extra-experimental-features "nix-command flakes")
features="kvm nixos-test big-parallel benchmark"
rm -f "$bake"

cleanup() {
  rm -f "$bake" "$bake.compact"
}
trap cleanup EXIT

echo "==> baking $out from $flake#$attr"
if command -v nix >/dev/null 2>&1; then
  result=$(nix "${nix_args[@]}" build "$flake#$attr" --no-link --print-out-paths)
  cp -L "$(find -L "$result" -name '*.qcow2' | head -n 1)" "$bake"
elif command -v docker >/dev/null 2>&1; then
  echo "==> host has no nix, baking via the nixos/nix container"
  docker_flake="$flake"
  mounts=(-v "$(pwd)":/workspace)
  if [ -d "$flake" ]; then
    mounts+=(-v "$(cd "$flake" && pwd)":/flake)
    docker_flake="path:/flake"
  fi
  docker run --rm --device /dev/kvm "${mounts[@]}" -v boite-bake-nix:/nix -w /workspace nixos/nix sh -c '
      set -eu
      result=$(nix --extra-experimental-features "nix-command flakes" build "$1" --no-link --print-out-paths \
        --option system-features "$2")
      cp -L "$(find -L "$result" -name "*.qcow2" | head -n 1)" "/workspace/$3"
      chmod 0644 "/workspace/$3"
    ' sh "$docker_flake#$attr" "$features" "$bake"
else
  echo "Error: neither nix nor docker found" >&2
  exit 1
fi

if [ ! -s "$bake" ]; then
  echo "Error: the build produced no qcow2 image" >&2
  exit 1
fi

chmod u+w "$bake"
qemu-img convert -O qcow2 -c "$bake" "$bake.compact"
mv "$bake.compact" "$bake"

sha=$(sha256sum "$bake" | awk '{print $1}')
mv "$bake" "$out"
echo "==> built $out ($(du -h "$out" | cut -f1))"
echo "SHA256: $sha"
echo "Pin it in cmd/qemu/distro.go (FluxImageSHA256) and upload $out to https://boite.facile.studio/flux.qcow2"
