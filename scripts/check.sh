#!/usr/bin/env sh

set -eu

GO_MODULES="."

mode="all"
case "${1:-}" in
--go-only) mode="go" ;;
--format) mode="format" ;;
"") ;;
*)
  echo "usage: $0 [--go-only|--format]" >&2
  exit 2
  ;;
esac

cd "$(git rev-parse --show-toplevel)"

if [ -z "${GO:-}" ]; then
  if [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/go" ]; then GO="$GOROOT/bin/go"; else GO=go; fi
fi
if [ -z "${GOFMT:-}" ]; then
  if [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/gofmt" ]; then GOFMT="$GOROOT/bin/gofmt"; else GOFMT=gofmt; fi
fi

if ! command -v "$GO" >/dev/null 2>&1 && [ ! -x "$GO" ]; then
  echo "check: no usable go ('$GO')" >&2
  exit 1
fi

if [ "$mode" = "format" ]; then
  for dir in $GO_MODULES; do
    (cd "$dir" && "$GO" fmt ./...)
  done
  exit 0
fi

status=0

for dir in $GO_MODULES; do
  echo "==> $dir"
  (
    cd "$dir" || exit 1
    s=0

    unformatted="$("$GOFMT" -l . | grep -v '^vendor/' || true)"
    if [ -n "$unformatted" ]; then
      echo "gofmt: the following files are not formatted (run 'sh scripts/check.sh --format'):"
      echo "$unformatted"
      s=1
    fi

    "$GO" vet ./... || s=1
    "$GO" test ./... || s=1

    exit "$s"
  ) || status=1
done

if command -v filet >/dev/null 2>&1; then
  echo "==> filet check"
  filet check . || status=1
fi

# QEMU toolchain checks (required for sandbox runtime)
echo "==> QEMU toolchain"
for bin in qemu-system-x86_64 qemu-img mcopy; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "check: $bin not found in PATH (required for sandbox runtime)" >&2
    status=1
  fi
done
# mkfs.fat lives in /sbin on Debian, which may not be on the non-login PATH
if ! command -v mkfs.fat >/dev/null 2>&1 && ! [ -x /sbin/mkfs.fat ]; then
  echo "check: mkfs.fat not found (required for sandbox runtime)" >&2
  status=1
fi

if [ "$status" -ne 0 ]; then
  echo "check failed"
else
    echo "check ok"
fi
exit "$status"
