# Input recording and replay are disabled in this fork. The original script is kept
# below, commented out, so it can be restored by uncommenting it.

# #!/bin/sh
# # E2E verification: record input through the Recordings window, replay it through the CLI, and compare behavior.
# #
# # Usage: sh verify-replay-via-cli.sh [--project-path <path>]
# #
# # Prerequisites:
# #   - Unity Editor running for the target project
# #   - PlayMode is not running because this script starts it
# #   - Record input with Window > Tuanjie CLI Loop > Recordings when prompted

# set -e

# PROJECT_PATH=""
# TLOOP_PATH="${TLOOP_BIN:-tloop}"
# SCENE_PATH="Assets/Scenes/InputReplayVerificationScene.unity"

# fail() {
#     echo "ERROR: $1" >&2
#     exit 1
# }

# while [ "$#" -gt 0 ]; do
#     case "$1" in
#         --project-path)
#             [ "$#" -ge 2 ] || fail "--project-path requires a value"
#             PROJECT_PATH=$2
#             shift 2
#             ;;
#         --tloop-path)
#             [ "$#" -ge 2 ] || fail "--tloop-path requires a value"
#             TLOOP_PATH=$2
#             shift 2
#             ;;
#         -h|--help)
#             echo "Usage: sh verify-replay-via-cli.sh [--project-path <path>] [--tloop-path <path>]"
#             exit 0
#             ;;
#         *)
#             fail "unknown option: $1"
#             ;;
#     esac
# done

# run_tloop() {
#     if [ -n "$PROJECT_PATH" ]; then
#         "$TLOOP_PATH" --project-path "$PROJECT_PATH" "$@"
#     else
#         "$TLOOP_PATH" "$@"
#     fi
# }

# RECORDING_LOG=".tloop/outputs/InputRecordings/recording-event-log.txt"
# REPLAY_LOG=".tloop/outputs/InputRecordings/replay-event-log.txt"

# if [ -n "$PROJECT_PATH" ]; then
#     RECORDING_LOG="$PROJECT_PATH/$RECORDING_LOG"
#     REPLAY_LOG="$PROJECT_PATH/$REPLAY_LOG"
# fi

# run_tloop_json() {
#     output=$(run_tloop "$@" 2>&1) || {
#         printf '%s\n' "$output" >&2
#         fail "tloop $* failed"
#     }
#     if printf '%s\n' "$output" | grep -Eq '"Success"[[:space:]]*:[[:space:]]*false'; then
#         printf '%s\n' "$output" >&2
#         fail "tloop $* returned success=false"
#     fi

#     printf '%s\n' "$output"
# }

# assert_json_result() {
#     json=$1
#     expected=$2
#     context=$3

#     actual=$(printf '%s\n' "$json" | sed -n 's/.*"Result"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
#     [ "$actual" = "$expected" ] || fail "$context: expected '$expected', got '$actual'"
# }

# wait_for_unity() {
#     i=0
#     while [ $i -lt 15 ]; do
#         if run_tloop get-logs --max-count 1 > /dev/null 2>&1; then
#             return 0
#         fi
#         sleep 2
#         i=$((i + 1))
#     done
#     echo "ERROR: Unity not responding"
#     exit 1
# }

# activate_for_record() {
#     json=$(run_tloop_json execute-dynamic-code --code '
# var cube = GameObject.Find("VerificationCube");
# if (cube == null) return "ERROR: VerificationCube not found";
# cube.SendMessage("ActivateForExternalControl");
# return "OK: activated for recording";
# ')
#     assert_json_result "$json" "OK: activated for recording" "Activate recording controller"
# }

# activate_for_replay() {
#     json=$(run_tloop_json execute-dynamic-code --code '
# var cube = GameObject.Find("VerificationCube");
# if (cube == null) return "ERROR: VerificationCube not found";
# cube.SendMessage("ActivateForExternalReplay");
# return "OK: activated for replay";
# ')
#     assert_json_result "$json" "OK: activated for replay" "Activate replay controller"
# }

# save_log() {
#     unity_path=$1
#     if command -v cygpath >/dev/null 2>&1; then
#         unity_path=$(cygpath -w "$1")
#     fi
#     escaped_path=$(printf '%s\n' "$unity_path" | sed 's/\\/\\\\/g; s/"/\\"/g')
#     rm -f "$1"
#     json=$(run_tloop_json execute-dynamic-code --code "
# var cube = GameObject.Find(\"VerificationCube\");
# if (cube == null) return \"ERROR: VerificationCube not found\";
# cube.SendMessage(\"SaveLog\", \"$escaped_path\");
# return \"OK: log saved\";
# ")
#     assert_json_result "$json" "OK: log saved" "Save event log"
#     [ -f "$1" ] || fail "Save event log did not create $1"
# }

# initialize_replay_scene() {
#     run_tloop control-play-mode --action Stop >/dev/null 2>&1 || true

#     json=$(run_tloop_json execute-dynamic-code --code "
# using UnityEditor.SceneManagement;
# using UnityEngine.SceneManagement;
# string scenePath = \"$SCENE_PATH\";
# if (SceneManager.GetActiveScene().path != scenePath)
# {
#     EditorSceneManager.OpenScene(scenePath, OpenSceneMode.Single);
# }
# return SceneManager.GetActiveScene().path;
# ")
#     assert_json_result "$json" "$SCENE_PATH" "Load replay verification scene"
# }

# echo ""
# echo "========================================="
# echo "  Input Replay E2E Verification"
# echo "========================================="

# # ---- Phase 1: Record human input through the Recordings window ----

# echo ""
# echo "[1/8] Loading replay verification scene..."
# initialize_replay_scene

# echo "[2/8] Starting PlayMode..."
# run_tloop_json control-play-mode --action Play >/dev/null
# echo "  Waiting for Unity..."
# sleep 6
# wait_for_unity

# echo "[3/8] Activating controller..."
# activate_for_record

# echo "[4/8] Recording input with the Recordings window..."
# echo ""
# echo "========================================="
# echo "  Open Window > Tuanjie CLI Loop > Recordings."
# echo "  Click Start Recording, play in the Game View, then click Stop Recording."
# echo ""
# echo "  WASD: move | Mouse: rotate"
# echo "  Left click: red | Right click: blue"
# echo "  Scroll: scale"
# echo ""
# echo "  Press ENTER here after the recording has stopped."
# echo "========================================="
# echo ""
# read -r _

# echo "[5/8] Saving recording event log..."
# save_log "$RECORDING_LOG"
# [ -s "$RECORDING_LOG" ] || fail "Recording event log is empty"

# # ---- Phase 2: Replay via CLI ----

# echo "[6/8] Restarting PlayMode..."
# run_tloop_json control-play-mode --action Stop >/dev/null
# sleep 3
# run_tloop_json control-play-mode --action Play >/dev/null
# echo "  Waiting for Unity..."
# sleep 6
# wait_for_unity

# echo "[7/8] Activating controller + starting replay via CLI..."
# activate_for_replay
# echo "  Starting replay of the latest Recordings window file..."
# REPLAY_RESULT=$(run_tloop replay-input --action start 2>&1) || true
# echo "  $REPLAY_RESULT"

# echo "  Waiting for replay to finish..."
# sleep 2
# waited=0
# while [ $waited -lt 60 ]; do
#     STATUS_RESULT=$(run_tloop replay-input --action status 2>&1) || true
#     playing=$(echo "$STATUS_RESULT" | grep -o '"IsReplaying": *[a-z]*' | sed 's/.*: *//')
#     if [ "$playing" = "false" ]; then
#         echo "  Replay completed."
#         break
#     fi
#     if [ $((waited % 5)) -eq 0 ]; then
#         progress=$(echo "$STATUS_RESULT" | grep -o '"Progress": *[0-9.]*' | sed 's/.*: *//')
#         echo "  Progress: ${progress:-...}"
#     fi
#     sleep 1
#     waited=$((waited + 1))
# done
# echo ""

# if [ $waited -ge 60 ]; then
#     echo "ERROR: Replay did not complete within 60s"
#     echo "  Last status: $STATUS_RESULT"
#     exit 1
# fi
# sleep 1

# echo "[8/8] Saving replay event log..."
# save_log "$REPLAY_LOG"
# [ -s "$REPLAY_LOG" ] || fail "Replay event log is empty"

# # ---- Phase 3: Compare ----

# echo ""
# echo "[Final] Comparing logs..."
# echo ""

# # Normalize frame numbers to relative (first event = frame 0).
# # CLI commands introduce variable delays, so absolute frame numbers
# # differ, but relative timing between events should be identical.
# normalize_frames() {
#     base=$(head -1 "$1" | sed 's/Frame \([0-9]*\):.*/\1/')
#     sed "s/Frame \([0-9]*\)/Frame \1/" "$1" | while IFS= read -r line; do
#         frame=$(echo "$line" | sed 's/Frame \([0-9]*\):.*/\1/')
#         rest=$(echo "$line" | sed 's/Frame [0-9]*: //')
#         echo "Frame $((frame - base)): $rest"
#     done
# }

# normalize_frames "$RECORDING_LOG" > "$RECORDING_LOG.norm"
# normalize_frames "$REPLAY_LOG" > "$REPLAY_LOG.norm"

# if diff "$RECORDING_LOG.norm" "$REPLAY_LOG.norm" > /dev/null 2>&1; then
#     lines=$(wc -l < "$RECORDING_LOG.norm" | tr -d ' ')
#     echo "========================================="
#     echo "  RESULT: MATCH ($lines events identical)"
#     echo "  Relative frame timing verified."
#     echo "========================================="
#     echo ""
#     rm -f "$RECORDING_LOG.norm" "$REPLAY_LOG.norm"
#     exit 0
# else
#     cnt=$(diff "$RECORDING_LOG.norm" "$REPLAY_LOG.norm" | grep -c '^[<>]' || true)
#     echo "========================================="
#     echo "  RESULT: MISMATCH ($cnt differences)"
#     echo "========================================="
#     echo ""
#     diff "$RECORDING_LOG.norm" "$REPLAY_LOG.norm" | head -20
#     echo ""
#     rm -f "$RECORDING_LOG.norm" "$REPLAY_LOG.norm"
#     exit 1
# fi
