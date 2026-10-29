# tloop cancel

Cancel the tloop command that holds the Editor's command slot, from another terminal or after the waiting command was given up on.

## When to Use

- Every command is rejected with `UNITY_SERVER_BUSY`, `tloop status` keeps reporting `Busy`, and the running command shows no sign of finishing.
  `MAIN_THREAD_BLOCKED` means the command never ran and released its hold on the slot, so `tloop cancel` is not needed for it (an `execute-dynamic-code` run sharing the slot may still hold it).
- A command you started does not return, for example an `execute-dynamic-code` snippet that awaits something that never completes.

## Usage

```bash
tloop cancel
tloop cancel --project-path <path>
```

It takes no other options. It answers while Tuanjie is busy and while a command blocks Tuanjie's main thread.

## Output

Returns JSON with:

- `Outcome`: what the cancel found
- `RunningToolName`, `RunningToolElapsedSeconds`: the command that held the slot; `null` for `NothingRunning`
- `Success`: `false` only for `NotCancellable`

| Outcome | Meaning | Exit code |
|---------|---------|-----------|
| `Cancelled` | The running command was asked to stop | 0 |
| `NothingRunning` | No command held the slot | 0 |
| `NotCancellable` | The running command cannot be cancelled and keeps running | 1 |

When Tuanjie cannot be reached, the error goes to stderr, nothing goes to stdout, and it exits 1. A Tuanjie package too old to support `tloop cancel` answers with a `UNITY_RPC_ERROR` naming the unknown command.

## What Cannot Be Cancelled

`NotCancellable` covers only these:

- `compile` that waits for the domain reload (the default)
- `run-tests --test-mode PlayMode --respect-enter-play-mode-settings`
- a tool started by code inside the Editor (an extension or an Editor window), not by the CLI

A default EditMode `run-tests` and `compile --no-wait-for-domain-reload` can be cancelled.

## What the Waiting Command Gets

Most commands fail with `ErrorCode` `TOOL_CANCELLED` (not retryable). `execute-dynamic-code` instead returns its usual cancelled result; the `tloop-execute-dynamic-code` skill's `references/cancellation.md` describes it and what happens to a snippet that keeps running.

`Cancelled` also resumes Play Mode when the Editor is paused, whether a pause point or a manual pause holds it — the same as when the waiting CLI is killed.

## After Cancelling

Run `tloop status` and wait for `Ready` before the next command. `Cancelled` means the command was asked to stop, not that it already stopped: a tool that ignores its cancellation token keeps the slot until it returns on its own.

The exception is an `execute-dynamic-code` snippet that still cannot return, for example because a command blocks Tuanjie's main thread. Polling `tloop status` never frees its slot, because status only reads the slot. Send a tool command other than `execute-dynamic-code` (such as `tloop get-logs`) instead: the slot is taken back 5 seconds after that command first finds it held, and the CLI keeps retrying the busy command through those 5 seconds, so the command then runs.
