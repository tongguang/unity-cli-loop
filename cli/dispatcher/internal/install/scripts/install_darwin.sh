InstallDir='{{INSTALL_DIR}}'
ExpectedUloopPath='{{EXPECTED_TLOOP_PATH}}'
PathBlockStart="# >>> tloop PATH >>>"
PathBlockEnd="# <<< tloop PATH <<<"

normalize_path() {
    printf '%s' "$1" | sed 's:/*$::'
}

path_contains_install_dir() {
    normalized_install_dir=$(normalize_path "$InstallDir")
    old_ifs=$IFS
    IFS=:
    for entry in ${PATH:-}; do
        if [ "$(normalize_path "$entry")" = "$normalized_install_dir" ]; then
            IFS=$old_ifs
            return 0
        fi
    done
    IFS=$old_ifs
    return 1
}

prepend_current_path() {
    if path_contains_install_dir; then
        return
    fi
    PATH="$InstallDir:${PATH:-}"
    export PATH
}

detect_user_shell_name() {
    shell_path=${SHELL:-}
    if [ -z "$shell_path" ]; then
        echo ""
        return
    fi
    echo "${shell_path##*/}"
}

detect_bash_profile_path() {
    # Linux terminal emulators start bash as a non-login shell, which reads
    # ~/.bashrc and never ~/.bash_profile.
    if [ "$(uname -s)" = "Linux" ]; then
        echo "$HOME/.bashrc"
        return
    fi
    if [ -f "$HOME/.bash_profile" ]; then
        echo "$HOME/.bash_profile"
        return
    fi
    if [ -f "$HOME/.bash_login" ]; then
        echo "$HOME/.bash_login"
        return
    fi
    if [ -f "$HOME/.profile" ]; then
        echo "$HOME/.profile"
        return
    fi
    echo "$HOME/.bash_profile"
}

detect_zsh_profile_path() {
    if [ -n "${ZDOTDIR:-}" ]; then
        echo "$ZDOTDIR/.zshrc"
        return
    fi
    echo "$HOME/.zshrc"
}

detect_fish_profile_path() {
    if [ -n "${XDG_CONFIG_HOME:-}" ]; then
        echo "$XDG_CONFIG_HOME/fish/config.fish"
        return
    fi
    echo "$HOME/.config/fish/config.fish"
}

quote_for_shell_profile() {
    printf "'"
    printf '%s' "$1" | sed "s/'/'\\\\''/g"
    printf "'"
}

resolve_profile_write_path() {
    profile_path=$1
    if [ ! -L "$profile_path" ]; then
        echo "$profile_path"
        return
    fi

    link_target=$(readlink "$profile_path" 2>/dev/null || true)
    if [ -z "$link_target" ]; then
        echo "$profile_path"
        return
    fi

    case "$link_target" in
        /*)
            echo "$link_target"
            ;;
        *)
            profile_dir=${profile_path%/*}
            if [ "$profile_dir" = "$profile_path" ]; then
                echo "$link_target"
                return
            fi
            echo "$profile_dir/$link_target"
            ;;
    esac
}

write_path_block() {
    profile_path=$1
    path_line=$2
    profile_write_path=$(resolve_profile_write_path "$profile_path")
    case "$profile_write_path" in
        */*)
            profile_dir=${profile_write_path%/*}
            [ -n "$profile_dir" ] || profile_dir=/
            ;;
        *)
            profile_dir=.
            ;;
    esac

    if [ "$profile_dir" != "." ] && [ "$profile_dir" != "/" ]; then
        if ! mkdir -p "$profile_dir"; then
            echo "Could not create shell profile directory: $profile_dir" >&2
            return 1
        fi
    fi

    tmp_path=$(mktemp "$profile_dir/.tloop_path.XXXXXX" 2>/dev/null || true)
    if [ -z "$tmp_path" ]; then
        echo "Could not create a temporary file for PATH setup." >&2
        return 1
    fi

    if [ -f "$profile_write_path" ]; then
        if ! awk -v start="$PathBlockStart" -v end="$PathBlockEnd" '
            $0 == start { skipping = 1; next }
            $0 == end { skipping = 0; next }
            !skipping { print }
        ' "$profile_write_path" > "$tmp_path"; then
            echo "Could not read existing shell profile: $profile_path" >&2
            rm -f "$tmp_path"
            return 1
        fi
    else
        : > "$tmp_path"
    fi

    {
        printf '\n%s\n' "$PathBlockStart"
        printf '%s\n' "$path_line"
        printf '%s\n' "$PathBlockEnd"
    } >> "$tmp_path"

    if mv "$tmp_path" "$profile_write_path"; then
        echo "Added $InstallDir to PATH in $profile_path. Open a new terminal to use it everywhere."
        return 0
    fi

    echo "Could not update shell profile: $profile_path" >&2
    rm -f "$tmp_path"
    return 1
}

configure_shell_path() {
    if [ -z "${HOME:-}" ]; then
        echo "Could not resolve HOME for shell PATH setup." >&2
        return 1
    fi

    shell_name=$(detect_user_shell_name)
    case "$shell_name" in
        zsh)
            write_path_block "$(detect_zsh_profile_path)" "export PATH=$(quote_for_shell_profile "$InstallDir"):\$PATH"
            ;;
        bash)
            write_path_block "$(detect_bash_profile_path)" "export PATH=$(quote_for_shell_profile "$InstallDir"):\$PATH"
            ;;
        fish)
            write_path_block "$(detect_fish_profile_path)" "fish_add_path --move $(quote_for_shell_profile "$InstallDir")"
            ;;
        *)
            echo "Add this directory to PATH in your shell profile:"
            echo "  $InstallDir"
            return 0
            ;;
    esac
}

infer_npm_prefix_from_tloop_path() {
    command_path=$1
    case "$command_path" in
        */bin/tloop|*/bin/tloop.exe)
            bin_dir=${command_path%/*}
            echo "${bin_dir%/bin}"
            ;;
        *)
            echo ""
            ;;
    esac
}

is_legacy_npm_tloop_path() {
    command_path=$1
    if [ -z "$command_path" ]; then
        return 1
    fi
    if [ "$command_path" = "$ExpectedUloopPath" ]; then
        return 1
    fi
    if [ -L "$command_path" ]; then
        link_target=$(readlink "$command_path" 2>/dev/null || true)
        case "$link_target" in
            *node_modules/tloop-cli*|*node_modules\\tloop-cli*) return 0 ;;
        esac
    fi
    if [ -f "$command_path" ] && grep -F "node_modules/tloop-cli" "$command_path" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

try_remove_legacy_npm_package() {
    legacy_tloop=$1
    legacy_prefix=""
    if ! is_legacy_npm_tloop_path "$legacy_tloop"; then
        return
    fi
    legacy_prefix=$(infer_npm_prefix_from_tloop_path "$legacy_tloop")
    if [ -z "$legacy_prefix" ]; then
        return
    fi
    if ! command -v npm >/dev/null 2>&1; then
        return
    fi
    if npm uninstall -g --prefix "$legacy_prefix" tloop-cli; then
        if ! is_legacy_npm_tloop_path "$legacy_tloop"; then
            echo "Removed legacy npm package: tloop-cli"
            return
        fi
    fi
}

try_remove_default_legacy_npm_package() {
    if ! command -v npm >/dev/null 2>&1; then
        return
    fi
    default_prefix=$(npm prefix -g 2>/dev/null || true)
    if [ -n "$default_prefix" ]; then
        default_tloop_path="$default_prefix/bin/tloop"
        if [ "$(normalize_path "$default_tloop_path")" = "$(normalize_path "$ExpectedUloopPath")" ]; then
            return
        fi
    fi
    npm uninstall -g tloop-cli >/dev/null 2>&1 || true
}

configure_legacy_cleanup() {
    legacy_tloop=$(command -v tloop 2>/dev/null || true)
    try_remove_legacy_npm_package "$legacy_tloop"
    try_remove_default_legacy_npm_package
}

report_path_shadowing() {
    resolved_tloop=$(command -v tloop 2>/dev/null || true)
    if [ -z "$resolved_tloop" ] || [ "$resolved_tloop" = "$ExpectedUloopPath" ]; then
        return
    fi
    echo "Installed tloop to $ExpectedUloopPath, but PATH resolves tloop to:"
    echo "  $resolved_tloop"
    echo "Move $InstallDir earlier in PATH, or remove the legacy installation if it owns that command."
}

shell_path_status=0
configure_shell_path || shell_path_status=$?
configure_legacy_cleanup
prepend_current_path
report_path_shadowing
exit "$shell_path_status"
