#!/bin/sh
set -e
# Keep keyboard and mouse overlays visible for layout adjustment.
# Holds Space key + Left mouse LongPress so both overlays stay on screen.
# Press Ctrl+C to stop.
#
# Usage: sh scripts/keep-overlays-visible.sh [--project-path <path>]

PROJECT_PATH=""
if [ "$1" = "--project-path" ] && [ -n "$2" ]; then
    PROJECT_PATH="$2"
fi

run_tloop() {
    if [ -n "$PROJECT_PATH" ]; then
        tloop "$@" --project-path "$PROJECT_PATH"
    else
        tloop "$@"
    fi
}

cleanup() {
    printf "\033[34m[overlay]\033[0m Releasing keys...\n"
    run_tloop simulate-keyboard --action KeyUp --key Space > /dev/null 2>&1 || true
    printf "\033[34m[overlay]\033[0m Stopping PlayMode...\n"
    run_tloop control-play-mode --action Stop > /dev/null 2>&1 || true
    printf "\033[34m[overlay]\033[0m Done.\n"
}
trap cleanup EXIT INT

printf "\033[34m[overlay]\033[0m Starting PlayMode...\n"
run_tloop control-play-mode --action Play > /dev/null
sleep 2

printf "\033[34m[overlay]\033[0m Holding Space + Left mouse LongPress...\n"
printf "\033[34m[overlay]\033[0m Press Ctrl+C to stop.\n"

run_tloop simulate-keyboard --action KeyDown --key Space > /dev/null
run_tloop simulate-mouse-input --action LongPress --x 400 --y 300 --button Left --duration 600
