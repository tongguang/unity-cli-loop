#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd "$(dirname "$0")/.." && pwd)

path_contains_dir() {
  expected_dir=$(normalize_path_dir "$1")
  old_ifs=$IFS
  IFS=:
  for path_dir in $PATH; do
    if [ "$(normalize_path_dir "$path_dir")" = "$expected_dir" ]; then
      IFS=$old_ifs
      return 0
    fi
  done
  IFS=$old_ifs
  return 1
}

normalize_path_dir() {
  path_dir="$1"
  while [ "$path_dir" != "/" ] && [ "${path_dir%/}" != "$path_dir" ]; do
    path_dir=${path_dir%/}
  done
  printf '%s\n' "$path_dir"
}

ensure_symlink_target() {
  link_path="$1"

  if [ -e "$link_path" ] && [ ! -L "$link_path" ]; then
    echo "Go CLI source checks passed and dist binaries were rebuilt."
    echo "Refusing to overwrite non-symlink global tloop: $link_path" >&2
    exit 1
  fi
}

update_tloop_link() {
  link_path="$1"

  if [ -L "$link_path" ]; then
    current_target=$(readlink "$link_path")
    echo "Updating global tloop symlink: $link_path -> $current_target"
  else
    echo "Creating global tloop symlink: $link_path"
  fi

  ln -sfn "$cli_path" "$link_path"
  echo "Global tloop now points at the rebuilt native CLI: $link_path -> $(readlink "$link_path")"
}

ensure_global_tloop_resolves_to_link() {
  resolved_tloop_path=$(command -v "$global_command_name" || true)

  if [ "$resolved_tloop_path" = "$global_tloop_path" ]; then
    return 0
  fi
  if [ -n "$extra_global_tloop_path" ] && [ "$resolved_tloop_path" = "$extra_global_tloop_path" ]; then
    return 0
  fi

  echo "Global $global_command_name symlink was updated, but shell resolution does not point at it." >&2
  echo "Resolved $global_command_name: ${resolved_tloop_path:-not found}" >&2
  echo "Expected $global_command_name: $global_tloop_path" >&2
  echo "Add $global_bin_dir to PATH or set TLOOP_GLOBAL_BIN_DIR to a directory earlier in PATH." >&2
  exit 1
}

"$ROOT_DIR/scripts/check-go-cli-source.sh"
"$ROOT_DIR/scripts/build-go-cli.sh"

cli_path=""
global_command_name="tloop"
existing_tloop_path=""
os=$(uname -s)
arch=$(uname -m)

case "$os:$arch" in
  Darwin:arm64 | Darwin:aarch64)
    cli_path="$ROOT_DIR/dist/darwin-arm64/tloop"
    ;;
  Darwin:x86_64 | Darwin:amd64)
    cli_path="$ROOT_DIR/dist/darwin-amd64/tloop"
    ;;
  MINGW*:x86_64 | MINGW*:amd64 | MSYS*:x86_64 | MSYS*:amd64 | CYGWIN*:x86_64 | CYGWIN*:amd64 | Windows_NT:x86_64 | Windows_NT:amd64)
    cli_path="$ROOT_DIR/dist/windows-amd64/tloop.exe"
    global_command_name="tloop.exe"
    ;;
esac

if [ -z "$cli_path" ]; then
  echo "Go CLI source checks passed and dist binaries were rebuilt."
  echo "No built native CLI is mapped for this platform: $os/$arch"
  exit 0
fi

if [ ! -x "$cli_path" ]; then
  echo "Native CLI was not built or is not executable: $cli_path" >&2
  exit 1
fi

global_bin_dir=""

if [ -n "${TLOOP_GLOBAL_BIN_DIR:-}" ]; then
  global_bin_dir="$TLOOP_GLOBAL_BIN_DIR"
elif command -v "$global_command_name" >/dev/null 2>&1; then
  existing_tloop_path=$(command -v "$global_command_name")
  global_bin_dir=$(dirname "$existing_tloop_path")
elif [ "$global_command_name" = "tloop.exe" ] && command -v tloop >/dev/null 2>&1; then
  existing_tloop_path=$(command -v tloop)
  global_bin_dir=$(dirname "$existing_tloop_path")
elif path_contains_dir "$HOME/.npm-global/bin"; then
  global_bin_dir="$HOME/.npm-global/bin"
elif path_contains_dir "$HOME/.local/bin"; then
  global_bin_dir="$HOME/.local/bin"
fi

if [ -z "$global_bin_dir" ]; then
  echo "Go CLI source checks passed and dist binaries were rebuilt."
  echo "No writable PATH directory was selected for global tloop." >&2
  echo "Set TLOOP_GLOBAL_BIN_DIR to the directory that should contain tloop." >&2
  exit 1
fi

mkdir -p "$global_bin_dir"
global_bin_dir=$(CDPATH= cd "$global_bin_dir" && pwd)
global_tloop_path="$global_bin_dir/$global_command_name"
extra_global_tloop_path=""

if [ "$global_command_name" = "tloop.exe" ] && [ -n "$existing_tloop_path" ] && [ "$existing_tloop_path" != "$global_tloop_path" ]; then
  existing_tloop_dir=$(dirname "$existing_tloop_path")
  if [ "$existing_tloop_dir" = "$global_bin_dir" ]; then
    extra_global_tloop_path="$existing_tloop_path"
  fi
fi

echo "Go CLI source checks passed and dist binaries were rebuilt."

ensure_symlink_target "$global_tloop_path"
if [ -n "$extra_global_tloop_path" ]; then
  ensure_symlink_target "$extra_global_tloop_path"
fi

update_tloop_link "$global_tloop_path"
if [ -n "$extra_global_tloop_path" ]; then
  update_tloop_link "$extra_global_tloop_path"
fi

ensure_global_tloop_resolves_to_link

"$global_command_name" --version
