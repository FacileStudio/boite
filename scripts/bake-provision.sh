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
  python3-dev genisoimage

# boite user with passwordless sudo
id boite >/dev/null 2>&1 || useradd -m -s /bin/zsh -c Boite -G sudo,docker boite
mkdir -p /etc/sudoers.d
printf '%s\n' 'boite ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/boite
chmod 440 /etc/sudoers.d/boite

# SSH host keys (the build appliance ships none)
rm -f /etc/ssh/ssh_host_*key /etc/ssh/ssh_host_*key.pub
ssh-keygen -A

# dotfiles
cat > /home/boite/.zshrc <<'ZRC'
#!/bin/zsh
export PATH="/usr/local/go/bin:$HOME/.cargo/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.bun/bin:$PATH"
export WORKSPACE=/workspace
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

# toolchain (network installs)
curl -fsSL https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh
curl -fsSL https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -
su - boite -c 'pip install --break-system-packages --upgrade pip'
su - boite -c 'pip install --break-system-packages black isort flake8 pytest pytest-cov mypy poetry ipython httpie'
su - boite -c 'curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y'
su - boite -c 'curl -fsSL https://bun.sh/install | bash'

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
blk=/dev/disk/by-label/BOITECFG
[ -e "$blk" ] || exit 1
mkdir -p /mnt/boitecfg
mount -o ro "$blk" /mnt/boitecfg
mkdir -p /home/boite/.ssh
cp /mnt/boitecfg/authorized_keys /home/boite/.ssh/authorized_keys
chown boite:boite /home/boite/.ssh/authorized_keys
chmod 600 /home/boite/.ssh/authorized_keys
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