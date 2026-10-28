#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd "$(dirname "$0")/.." && pwd)
ACTION="${1:-status}"

command_name="tloop"

normalize_path_dir() {
  path_dir="$1"
  while [ "$path_dir" != "/" ] && [ "${path_dir%/}" != "$path_dir" ]; do
    path_dir=${path_dir%/}
  done
  printf '%s\n' "$path_dir"
}

select_local_tloop() {
  os_name=$(uname -s)
  arch_name=$(uname -m)

  case "$os_name:$arch_name" in
    Darwin:arm64|Darwin:aarch64)
      local_tloop="$ROOT_DIR/dist/darwin-arm64/tloop"
      ;;
    Darwin:x86_64|Darwin:amd64)
      local_tloop="$ROOT_DIR/dist/darwin-amd64/tloop"
      ;;
    *)
      echo "No built local tloop binary for $os_name $arch_name." >&2
      exit 1
      ;;
  esac

  if [ ! -x "$local_tloop" ]; then
    echo "Local tloop binary is missing or not executable: $local_tloop" >&2
    exit 1
  fi
}

path_contains_dir() {
  candidate_dir=$(normalize_path_dir "$1")
  old_ifs=$IFS
  IFS=:
  for path_dir in $PATH; do
    normalized_path_dir=$(normalize_path_dir "$path_dir")
    if [ "$normalized_path_dir" = "$candidate_dir" ]; then
      IFS=$old_ifs
      return 0
    fi
  done
  IFS=$old_ifs
  return 1
}

select_global_bin_dir() {
  if [ "${TLOOP_GLOBAL_BIN_DIR:-}" ]; then
    global_bin_dir=$TLOOP_GLOBAL_BIN_DIR
    return
  fi

  existing_tloop=$(command -v "$command_name" 2>/dev/null || true)
  if [ "$existing_tloop" ]; then
    global_bin_dir=$(dirname "$existing_tloop")
    return
  fi

  npm_global_bin="$HOME/.npm-global/bin"
  if path_contains_dir "$npm_global_bin" || [ -e "$npm_global_bin/$command_name" ]; then
    global_bin_dir=$npm_global_bin
    return
  fi

  local_bin="$HOME/.local/bin"
  if path_contains_dir "$local_bin"; then
    global_bin_dir=$local_bin
    return
  fi

  echo "No PATH directory found for global tloop. Set TLOOP_GLOBAL_BIN_DIR." >&2
  exit 1
}

print_status() {
  echo "Global tloop path: $global_tloop"
  if [ -L "$global_tloop" ]; then
    echo "Global tloop symlink target: $(readlink "$global_tloop")"
  elif [ -e "$global_tloop" ]; then
    echo "Global tloop is a regular file or directory."
  else
    echo "Global tloop does not exist."
  fi

  if [ -e "$backup_tloop" ] || [ -L "$backup_tloop" ]; then
    echo "Backup path: $backup_tloop"
  else
    echo "Backup path: none"
  fi

  echo "Local branch tloop: $local_tloop"
  if [ -x "$global_tloop" ]; then
    echo "Selected global tloop version:"
    "$global_tloop" --version || true
  fi

  if command -v "$command_name" >/dev/null 2>&1; then
    echo "Resolved tloop: $(command -v "$command_name")"
  fi
}

link_local_tloop() {
  mkdir -p "$global_bin_dir"

  if [ -L "$global_tloop" ] && [ "$(readlink "$global_tloop")" = "$local_tloop" ]; then
    echo "Global tloop already points to this checkout."
    print_status
    return
  fi

  if [ -e "$global_tloop" ] || [ -L "$global_tloop" ]; then
    if [ -e "$backup_tloop" ] || [ -L "$backup_tloop" ]; then
      echo "Backup already exists, refusing to overwrite it: $backup_tloop" >&2
      echo "Run '$0 restore' first or move the backup manually." >&2
      exit 1
    fi

    mv "$global_tloop" "$backup_tloop"
    echo "Backed up existing global tloop to $backup_tloop"
  fi

  ln -s "$local_tloop" "$global_tloop"
  print_status
}

restore_global_tloop() {
  if [ ! -e "$backup_tloop" ] && [ ! -L "$backup_tloop" ]; then
    echo "No backup found: $backup_tloop" >&2
    exit 1
  fi

  if [ -L "$global_tloop" ]; then
    current_target=$(readlink "$global_tloop")
    if [ "$current_target" != "$local_tloop" ]; then
      echo "Refusing to remove unexpected tloop symlink: $global_tloop -> $current_target" >&2
      exit 1
    fi

    rm "$global_tloop"
  elif [ -e "$global_tloop" ]; then
    echo "Refusing to overwrite unexpected global tloop: $global_tloop" >&2
    exit 1
  fi

  mv "$backup_tloop" "$global_tloop"
  print_status
}

usage() {
  echo "Usage: $0 link|restore|status" >&2
  exit 1
}

select_local_tloop
select_global_bin_dir

global_tloop="$global_bin_dir/$command_name"
backup_tloop="$global_bin_dir/$command_name.before-local-link"

case "$ACTION" in
  link)
    link_local_tloop
    ;;
  restore)
    restore_global_tloop
    ;;
  status)
    print_status
    ;;
  *)
    usage
    ;;
esac
