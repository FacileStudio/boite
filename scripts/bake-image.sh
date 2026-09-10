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

if ! qemu-img info "$bake" | grep -q "file format: qcow2"; then
  echo "virt-builder did not emit qcow2; converting" >&2
  qemu-img convert -O qcow2 -c "$bake" "$bake.qcow2"
  mv "$bake.qcow2" "$bake"
fi

sha=$(sha256sum "$bake" | awk '{print $1}')
mv "$bake" "$out"
echo "==> built $out"
echo "SHA256: $sha"
echo "Repin cmd/qemu/instance.go:"
echo "  BaseImageName    = \"$out\""
echo "  BaseImageSHA256  = \"$sha\""
echo "  BaseImageURL     = <where the image is published>"