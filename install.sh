#!/usr/bin/env bash
#
# Facile Studio installer. This is a shim by design: the installation logic
# lives in `facile`, the suite installer, so the suite has exactly one
# implementation of it instead of one copy per repo. Canonical shape in
# ~/.mycelium/memory/standards/cli.md.
#
# Equivalent, once facile is on your PATH:
#   facile install boite
#

set -euo pipefail

TOOL="boite"
BOOTSTRAP="https://get.facile.studio"
BOOTSTRAP_FALLBACK="https://raw.githubusercontent.com/FacileStudio/facile/main/install.sh"

bootstrap_facile() {
  command -v curl >/dev/null 2>&1 ||
    { printf '\033[31m✗\033[0m curl not found — install curl first\n' >&2; exit 1; }
  curl -fsSL "$BOOTSTRAP" | bash ||
    curl -fsSL "$BOOTSTRAP_FALLBACK" | bash
  export PATH="${FACILE_BIN_DIR:-$HOME/.local/bin}:$PATH"
}

main() {
  command -v facile >/dev/null 2>&1 || bootstrap_facile
  exec facile install "$TOOL" "$@"
}

main "$@"