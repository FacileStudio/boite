FROM node:22-bookworm-slim AS base

# Build args for user customization
ARG USER_NAME=devuser
ARG USER_UID=1000
ARG USER_GID=1000

# System tools needed for development
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    git curl unzip ca-certificates bash-completion fzf jq wget make tmux \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install mise for tool version management
RUN curl https://mise.run | sh
ENV PATH="/root/.local/bin:$PATH"

# Install Go
RUN curl -L https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -

# Install Rust via rustup
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:$PATH"

# Install Bun
RUN npm install -g bun@1.3.14

# Install skatos from GitHub (Rust-based env manager)
RUN curl -fsSL https://raw.githubusercontent.com/saravenpi/skatos/main/install.sh | bash

# Configure git globally
RUN git config --global init.defaultBranch main \
    && git config --global safe.directory '*'

# Create dev user and workspace
RUN groupadd --gid ${USER_GID} ${USER_NAME} || true
RUN useradd --uid ${USER_UID} --gid ${USER_GID} --create-home --shell /bin/bash ${USER_NAME}
WORKDIR /workspace
RUN chown ${USER_NAME}:${USER_NAME} /workspace
ENV HOME=/workspace USER=${USER_NAME}

# Copy mise configuration
COPY mise.toml /home/${USER_NAME}/.config/mise.toml
RUN chown ${USER_NAME}:${USER_NAME} /home/${USER_NAME}/.config/mise.toml

# Switch to non-root for dev work
USER ${USER_NAME}

# Install zsh with minimal config
RUN curl -sL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -o /tmp/install.sh \
    && sh /tmp/install.sh 2>&1 | tail -1
ENV SHELL=/bin/zsh

# Copy user's zsh config
COPY .zshrc_local /home/${USER_NAME}/.zshrc

# Copy tmux config
COPY .tmux.conf_local /home/${USER_NAME}/.tmux.conf
RUN chown ${USER_NAME}:${USER_NAME} /home/${USER_NAME}/.tmux.conf

# Default command is zsh shell
CMD ["/bin/zsh"]
