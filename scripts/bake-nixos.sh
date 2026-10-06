#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

out="${1:-boite-nixos.qcow2}"
bake="boite-nixos.baked.qcow2"
config_file="$(pwd)/nix/boite-configuration.nix"
rm -f "$bake"

cleanup() {
  rm -f "$bake"
  rm -rf ./result
}
trap cleanup EXIT

echo "==> baking $out with NixOS"
image_path=""
if command -v nixos-generate >/dev/null 2>&1; then
  image_path=$(nixos-generate -f qcow -c "$config_file")
elif command -v nixos-rebuild >/dev/null 2>&1; then
  nixos-rebuild build-image --image-variant qcow -I nixos-config="$config_file"
  image_path=$(find -L ./result -name "*.qcow2" | head -n 1)
elif command -v nix-build >/dev/null 2>&1; then
  nix-build '<nixpkgs/nixos>' -A config.system.build.qcow2 -I nixos-config="$config_file"
  image_path=$(find -L ./result -name "*.qcow2" | head -n 1)
elif command -v docker >/dev/null 2>&1; then
  echo "==> host has no nix toolchain, baking via docker nixos/nix container"
  docker run --rm --privileged --device /dev/kvm -v "$(pwd)":/workspace -w /workspace nixos/nix \
    nix-shell -p nixos-generators --run '
      target=$(nixos-generate --system x86_64-linux -f qcow -c /workspace/nix/boite-configuration.nix --option system-features "kvm benchmark big-parallel nixos-test uid-range")
      if [ -f "$target" ]; then
        cp "$target" /workspace/boite-nixos.baked.qcow2
      elif [ -f "$target/nixos.qcow2" ]; then
        cp "$target/nixos.qcow2" /workspace/boite-nixos.baked.qcow2
      else
        cp "$target"/*.qcow2 /workspace/boite-nixos.baked.qcow2
      fi
    '
  image_path="$bake"
else
  echo "Error: neither nixos-generate, nixos-rebuild, nix-build, nor docker found" >&2
  exit 1
fi

if [ -z "$image_path" ] || [ ! -f "$image_path" ]; then
  echo "Error: image generation failed to produce a qcow2 artifact" >&2
  exit 1
fi

if [ "$image_path" != "$bake" ]; then
  cp "$image_path" "$bake"
fi

apparent_bytes=$(stat -c %s "$bake" 2>/dev/null || stat -f %z "$bake")
apparent_gb=$((apparent_bytes / 1073741824))
if [ "${apparent_gb:-0}" -gt 4 ]; then
  echo "==> compacting apparent ${apparent_gb}G -> real $(du -h "$bake" | cut -f1)"
  qemu-img convert -O qcow2 -c "$bake" "$bake.compact"
  mv "$bake.compact" "$bake"
else
  qemu-img convert -O qcow2 -c "$bake" "$bake.compact"
  mv "$bake.compact" "$bake"
fi

sha=$(sha256sum "$bake" | awk '{print $1}')
mv "$bake" "$out"
echo "==> built $out ($(du -h "$out" | cut -f1))"
echo "SHA256: $sha"
