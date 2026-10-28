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
prepend_current_path
report_path_shadowing
exit "$shell_path_status"
