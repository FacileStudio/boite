#!/usr/bin/env bash
# Build the baked boite base image with virt-builder. Run by a maintainer when
# Debian or the toolchain changes; boite never calls this at runtime. After a
# successful build it prints the SHA256 to repin into cmd/qemu/instance.go and
# the URL where the image must be published.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

out="${1:-boite.qcow2}"
bake="boite.baked.qcow2"
rm -f "$bake"

echo "==> baking $out (this downloads Debian, installs the toolchain, and may take several minutes)"
echo "    (run 'virt-builder --update' first if the index is stale)"
virt-builder debian-13 \
  --output "$bake" \
  --format qcow2 \
  --size 20G \
  --run "$(pwd)/scripts/bake-provision.sh"

# virt-builder emits a sparse qcow2 whose on-disk size (du) is small but whose
# apparent size (stat %s -- what nginx serves as Content-Length and clients
# download) is the full 20G virtual volume. That would make the artifact a
# 20G download, so compact with -c when the apparent size is too large.
apparent_bytes=$(stat -c %s "$bake" 2>/dev/null || stat -f %z "$bake")
apparent_gb=$((apparent_bytes / 1073741824))
if [ "${apparent_gb:-0}" -gt 4 ]; then
  echo "==> compacting apparent ${apparent_gb}G -> real $(du -h "$bake" | cut -f1)"
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