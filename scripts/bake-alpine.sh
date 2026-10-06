#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

out="${1:-boite-alpine.qcow2}"
bake="boite-alpine.baked.qcow2"
rm -f "$bake"

temp_dir=$(mktemp -d)

cleanup() {
  rm -rf "$temp_dir"
  rm -f "$bake"
}
trap cleanup EXIT

make_vm_image=$(command -v alpine-make-vm-image || true)
if [ -z "$make_vm_image" ]; then
  curl -fsSL https://raw.githubusercontent.com/alpinelinux/alpine-make-vm-image/v0.12.0/alpine-make-vm-image -o "$temp_dir/alpine-make-vm-image"
  chmod +x "$temp_dir/alpine-make-vm-image"
  make_vm_image="$temp_dir/alpine-make-vm-image"
fi

cp scripts/firstboot-alpine.sh "$temp_dir/firstboot-alpine.sh"
chmod +x "$temp_dir/firstboot-alpine.sh"

tiroir_version="0.2.0"
curl -fsSL "https://github.com/FacileStudio/tiroir/releases/download/v${tiroir_version}/tiroir_${tiroir_version}_linux_amd64.tar.gz" -o "$temp_dir/tiroir.tar.gz"
tar -C "$temp_dir" -xzf "$temp_dir/tiroir.tar.gz" tiroir
chmod 0755 "$temp_dir/tiroir"
rm -f "$temp_dir/tiroir.tar.gz"

cat > "$temp_dir/setup.sh" <<'SETUP'
#!/bin/sh
set -eu

adduser -D -s /bin/sh boite
addgroup boite wheel 2>/dev/null || true

mkdir -p /etc/sudoers.d
printf '%s\n' 'boite ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/boite
chmod 0440 /etc/sudoers.d/boite

mkdir -p /etc/doas.d
printf '%s\n' 'permit nopass boite as root' > /etc/doas.d/doas.conf
printf '%s\n' 'permit nopass boite as root' > /etc/doas.conf
chmod 0600 /etc/doas.conf /etc/doas.d/doas.conf 2>/dev/null || true

rc-update add sshd default
ssh-keygen -A

cat > /etc/network/interfaces <<'NET'
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet dhcp
NET

rc-update add networking boot

mkdir -p /workspace /home/boite/.ssh /var/lib/boite
chown -R boite:boite /workspace /home/boite
chmod 0700 /home/boite/.ssh
chmod 0750 /home/boite

cat > /home/boite/.profile <<'PROF'
export PATH="/usr/local/bin:$HOME/.local/bin:$PATH"
export WORKSPACE=/workspace
if command -v tiroir >/dev/null 2>&1; then
  eval "$(tiroir export)"
fi
PROF
chown boite:boite /home/boite/.profile
chmod 0644 /home/boite/.profile

install -D -m 0755 /mnt/tiroir /usr/local/bin/tiroir
install -D -m 0755 /mnt/firstboot-alpine.sh /etc/init.d/boite-firstboot
mkdir -p /usr/local/lib/boite
install -m 0755 /mnt/firstboot-alpine.sh /usr/local/lib/boite/firstboot.sh
rc-update add boite-firstboot default
SETUP
chmod +x "$temp_dir/setup.sh"

sudo_cmd=""
if [ "$(id -u)" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then
    sudo_cmd="sudo"
  else
    echo "Error: alpine-make-vm-image must be run as root" >&2
    exit 1
  fi
fi

echo "==> baking $out with alpine-make-vm-image"
$sudo_cmd "$make_vm_image" \
  --image-format qcow2 \
  --image-size 20G \
  --packages "openssh sudo doas curl tar e2fsprogs util-linux ca-certificates" \
  --script-chroot \
  "$bake" \
  "$temp_dir/setup.sh"

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
