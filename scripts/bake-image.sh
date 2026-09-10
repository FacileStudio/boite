#!/usr/bin/env bash
# Build the baked boite base image with virt-builder. Run by a maintainer when
# Debian or the toolchain changes; boite never calls this at runtime. After a
# successful build it prints the SHA256 to repin into cmd/qemu/instance.go and
# the URL where the image must be published.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

out="${1:-boite-base.qcow2}"
bake="boite-base.baked.qcow2"
rm -f "$bake"

echo "==> baking $out (this downloads Debian, installs the toolchain, and may take several minutes)"
echo "    (run 'virt-builder --update' first if the index is stale)"
virt-builder debian-13 \
  --output "$bake" \
  --format qcow2 \
  --size 20G \
  --run "$(pwd)/scripts/bake-provision.sh"

size_gb=$(du -m "$bake" | awk '{print int($1)}')
apparent_gb=$(qemu-img info "$bake" | awk '/^virtual size:/{print $3}')
if [ "${size_gb:-0}" -gt 4000 ]; then
  # virt-builder emits a sparse qcow2 whose apparent size is the full virtual
  # volume; nginx serves the apparent size, so a 20G-virtual image would be
  # downloaded as 20G. Compact with -c so the artifact is only its real ~2G.
  echo "==> compacting $(du -h "$bake" | cut -f1) -> real size"
  qemu-img convert -O qcow2 -c "$bake" "$bake.compact"
  mv "$bake.compact" "$bake"
fi

sha=$(sha256sum "$bake" | awk '{print $1}')
mv "$bake" "$out"
echo "==> built $out ($(du -h "$out" | cut -f1))"
echo "SHA256: $sha"
echo "Repin cmd/qemu/instance.go:"
echo "  BaseImageName    = \"$out\""
echo "  BaseImageSHA256  = \"$sha\""
echo "  BaseImageURL     = <where the image is published>"