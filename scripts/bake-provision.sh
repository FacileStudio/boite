#!/usr/bin/env bash
# Runs inside the virt-builder appliance while the baked base image is being
# built. Installs the toolchain, creates the boite user, regenerates SSH host
# keys, lays down dotfiles, installs the firstboot oneshot and purges
# cloud-init. This is build-time provisioning, so nothing here is per-instance:
# the only per-instance input (the SSH public key) is installed at runtime by
# scripts/firstboot.sh from the config ISO.
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
export LANG=C.UTF-8

apt-get update -y

apt-get install -y --no-install-recommends \
  git curl unzip ca-certificates bash-completion fzf jq tmux wget make \
  build-essential zsh neovim starship docker.io gh tree htop btop zoxide \
  eza bat ripgrep prettyping fastfetch python3 python3-pip python3-venv \
  python3-dev genisoimage sudo

# boite user with passwordless sudo
id boite >/dev/null 2>&1 || useradd -m -s /bin/zsh -c Boite -G sudo,docker boite
mkdir -p /etc/sudoers.d
printf '%s\n' 'boite ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/boite
chmod 440 /etc/sudoers.d/boite

# Hostname. The virt-builder template leaves "unassigned-hostname"; every
# sandbox should be named boite so the shell prompt and hostname are stable.
echo boite > /etc/hostname
hostnamectl set-hostname boite 2>/dev/null || true

# Networking. The virt-builder debian-13 template predicts a "ens2" NIC, but
# under QEMU the virtio NIC is renamed eth0 -> ens3 (and sometimes ens18),
# so the template's "allow-hotplug ens2" never activates and the guest gets
# no address -- which silently breaks the SSH hostfwd (slirp accepts the host
# TCP but cannot route to a guest without an IP). Bring up whatever Ethernet
# interface is actually present via systemd-networkd with a wildcard match.
apt-get install -y --no-install-recommends systemd-networkd >/dev/null 2>&1 || true
mkdir -p /etc/systemd/network
cat > /etc/systemd/network/10-dhcp.network <<'NET'
[Match]
Name=e*

[Network]
DHCP=yes
NET
systemctl enable systemd-networkd 2>/dev/null || true

# SSH host keys (the build appliance ships none)
rm -f /etc/ssh/ssh_host_*key /etc/ssh/ssh_host_*key.pub
ssh-keygen -A

# Ensure sshd listens on TCP port 22 (boite reaches the guest over hostfwd).
# systemd-ssh-generator can leave only unix-local/vsock seats active, which
# slirp's hostfwd cannot reach; socket-activate the TCP listener explicitly.
systemctl enable ssh.socket 2>/dev/null || true
systemctl enable ssh.service 2>/dev/null || true

# dotfiles
cat > /home/boite/.zshrc <<'ZRC'
#!/bin/zsh
export PATH="/usr/local/go/bin:$HOME/.cargo/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.bun/bin:$PATH"
export WORKSPACE=/workspace
if command -v tiroir >/dev/null 2>&1; then
  eval "$(tiroir export)"
fi
if command -v mise >/dev/null 2>&1; then
  eval "$(mise activate zsh)"
elif [ -x /usr/local/bin/mise ]; then
  eval "$(/usr/local/bin/mise activate zsh)"
fi
command -v starship >/dev/null 2>&1 && eval "$(starship init zsh)"
if command -v eza >/dev/null 2>&1; then
  alias ls="eza --color=always --git --icons=always --no-user"
  alias tn="eza --tree --color=always --git --icons=always --no-user"
elif command ls --color=auto . >/dev/null 2>&1; then
  alias ls="ls --color=auto"
fi
alias ll='ls -l'
alias la='ls -a'
alias lal='ls -la'
alias l='ls --color=auto'
alias grep='grep --color=auto'
alias c='clear'
alias q='exit'
alias :q='exit'
alias ..='cd ..'
alias ...='cd ../..'
alias gs='git status'
alias ga='git add'
alias gc='git clone'
alias gd='git diff'
alias gb='git branch'
alias gco='git checkout'
alias ta='tmux attach -t'
alias tn='tmux new -s'
alias tl='tmux ls'
autoload -Uz compinit && compinit -C
ZRC
chown boite:boite /home/boite/.zshrc
chmod 644 /home/boite/.zshrc

# Non-interactive `sh -lc` never reads rc (boite exec paints the env itself
# via WithEnv), but a shell started as bash still should see managed keys.
cat > /home/boite/.bashrc <<'BRC'
#!/bin/bash
export PATH="/usr/local/bin:$HOME/.local/bin:$PATH"
if command -v tiroir >/dev/null 2>&1; then
  eval "$(tiroir export)"
fi
BRC
chown boite:boite /home/boite/.bashrc
chmod 644 /home/boite/.bashrc

cat > /home/boite/.tmux.conf <<'TMC'
set -g default-terminal "screen-256color"
set -ag terminal-overrides ",xterm-256color:RGB"
set -g mouse on
set -g default-shell /bin/zsh
set -g focus-events on
set -g escape-time 0
set -g history-limit 50000
set-window-option -g mode-keys vi
bind | split-window -h -c "#{pane_current_path}"
bind - split-window -v -c "#{pane_current_path}"
bind c new-window -c "#{pane_current_path}"
bind h select-pane -L
bind j select-pane -D
bind k select-pane -U
bind l select-pane -R
bind r source-file ~/.tmux.conf \; display-message "Config reloaded!"
TMC
chown boite:boite /home/boite/.tmux.conf
chmod 644 /home/boite/.tmux.conf

touch /home/boite/.hushlogin
chown boite:boite /home/boite/.hushlogin
rm -f /etc/legal /etc/motd

# toolchain (network installs). Every tool is pinned to a specific version so
# the baked image is reproducible: the SHA of the produced image travels with a
# known toolchain, not whatever "latest" was on bake day.
export MISE_VERSION=2026.9.4
export RUST_TOOLCHAIN=1.98.1
export BUN_VERSION=1.4.2
export TIROIR_VERSION=0.2.0

curl -fsSL https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh
curl -fsSL https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -
su - boite -c 'pip install --break-system-packages --upgrade pip'
su - boite -c 'pip install --break-system-packages black isort flake8 pytest pytest-cov mypy poetry ipython httpie'
su - boite -c "export RUST_TOOLCHAIN=$RUST_TOOLCHAIN; curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain \"\$RUST_TOOLCHAIN\""
su - boite -c "export BUN_VERSION=$BUN_VERSION; curl -fsSL \"https://github.com/oven-sh/bun/releases/download/bun-v\$BUN_VERSION/bun-linux-x64.zip\" -o /tmp/bun.zip && unzip -o /tmp/bun.zip -d /tmp/bun-extract && mkdir -p ~/.bun/bin && install /tmp/bun-extract/bun-linux-x64/bun ~/.bun/bin/bun && rm -rf /tmp/bun.zip /tmp/bun-extract"

# tiroir is the encrypted env store the guest shells load at login and boite
# manages from the host. Pinned to the same release boite consumes (go.mod
# replace points at git HEAD, which is v0.2.0).
curl -fsSL "https://github.com/FacileStudio/tiroir/releases/download/v$TIROIR_VERSION/tiroir_${TIROIR_VERSION}_linux_amd64.tar.gz" -o /tmp/tiroir.tar.gz
tar -C /usr/local/bin -xzf /tmp/tiroir.tar.gz tiroir
chmod 0755 /usr/local/bin/tiroir
rm -f /tmp/tiroir.tar.gz

su - boite -c 'git config --global init.defaultBranch main'
su - boite -c 'git config --global safe.directory "*"'

mkdir -p /workspace
chown -R boite:boite /workspace
ln -sfn /workspace /home/boite/workspace

mkdir -p /home/boite/.ssh
chown -R boite:boite /home/boite/.ssh
chmod 700 /home/boite/.ssh

su - boite -c 'mkdir -p ~/.local/share/mise/migrations ~/.cache/starship'
su - boite -c 'chown -R boite:boite ~/.local/share/mise ~/.cache/starship'
chown boite:boite /home/boite
chmod 750 /home/boite

# firstboot oneshot
mkdir -p /usr/local/lib/boite
cat > /usr/local/lib/boite/firstboot.sh <<'FB'
#!/bin/sh
set -e
marker=/var/lib/boite/firstboot.done
[ -f "$marker" ] && exit 0

mkdir -p /mnt/boitecfg
blk=
for attempt in 1 2 3 4 5 6; do
  for cand in /dev/disk/by-label/BOITECFG /dev/vdb /dev/vdc /dev/sdb /dev/sdc /dev/sr0 /dev/sr1; do
    if [ -b "$cand" ] && mount -o ro "$cand" /mnt/boitecfg 2>/dev/null && [ -f /mnt/boitecfg/authorized_keys ]; then
      blk=$cand
      break
    fi
    umount /mnt/boitecfg 2>/dev/null || true
  done
  [ -n "$blk" ] && break
  sleep 2
done
[ -n "$blk" ] && [ -f /mnt/boitecfg/authorized_keys ] || exit 1

mkdir -p /home/boite/.ssh /var/lib/boite
cp /mnt/boitecfg/authorized_keys /home/boite/.ssh/authorized_keys
chown boite:boite /home/boite/.ssh/authorized_keys
chmod 600 /home/boite/.ssh/authorized_keys
if [ -f /mnt/boitecfg/.tiroir ] && [ -f /mnt/boitecfg/.tiroir.key ]; then
  cp /mnt/boitecfg/.tiroir /mnt/boitecfg/.tiroir.key /home/boite/
  chown boite:boite /home/boite/.tiroir /home/boite/.tiroir.key
  chmod 600 /home/boite/.tiroir /home/boite/.tiroir.key
fi
umount /mnt/boitecfg
touch "$marker"
FB
chmod 0755 /usr/local/lib/boite/firstboot.sh
cat > /etc/systemd/system/boite-firstboot.service <<'UNIT'
[Unit]
Description=Boite firstboot provisioning
ConditionPathExists=!/var/lib/boite/firstboot.done
Before=sshd.service

[Service]
Type=oneshot
ExecStart=/usr/local/lib/boite/firstboot.sh
RemainAfterExit=no

[Install]
WantedBy=multi-user.target
UNIT
systemctl enable boite-firstboot.service

# cloud-init is dead weight on a baked image; purge it so its absence is verifiable
if dpkg -s cloud-init >/dev/null 2>&1; then
  apt-get purge -y cloud-init cloud-images 2>/dev/null || true
fi
apt-get autoremove -y || true
apt-get clean || true