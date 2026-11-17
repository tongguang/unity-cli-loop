---
name: tloop-status
description: "Check whether the Tuanjie Editor for this project can take a tloop command now, without side effects; also covers `tloop cancel` for stopping the command it is running and `tloop get-dialogs` / `tloop press-dialog-button` for listing the dialogs that block it and pressing one of their buttons. Use before sending commands when Tuanjie may be starting, compiling, busy, or closed, and when BUSY does not clear or a command does not finish."
---

# tloop status

Report whether the Tuanjie Editor for this project can take a tloop command now, without changing anything.

## Usage

```bash
tloop status
tloop status --project-path <path>
```

## States

The states are checked in this order, and the first one that matches is reported. A dialog found in Tuanjie's windows replaces `Ready`, `Compiling`, `ImportingAssets`, `Busy`, `MainThreadBlocked`, or `ServerUnavailable`.

| State | Meaning | Exit code |
|-------|---------|-----------|
| `BlockedByDialog` | A modal dialog is open and Tuanjie's main thread waits for it; how it is found and how to answer it are in Notes | 1 |
| `Busy` | Tuanjie answered, and a tloop command is running and holds the command slot | 1 |
| `Starting` | Tuanjie answered but has not recorded its play and compile state yet | 1 |
| `MainThreadBlocked` | Tuanjie answered, but its main thread has not run for 5 seconds or more, even when asked again after being woken | 1 |
| `Compiling` | Tuanjie answered and is compiling scripts | 1 |
| `ImportingAssets` | Tuanjie answered and is importing assets | 1 |
| `Ready` | Tuanjie answered and none of the above applies | 0 |
| `NotResponding` | Connected to Tuanjie, but it did not answer within 5 seconds | 1 |
| `ServerUnavailable` | The tloop server refused or dropped the connection, and Tuanjie is running for this project | 1 |
| `NotRunning` | The tloop server refused or dropped the connection, and no Tuanjie Editor is running for this project | 1 |
| `Unreachable` | The tloop server refused or dropped the connection, and the Tuanjie process list could not be read | 1 |

Some failures are errors, not states: the operating system refusing the connection (as a sandbox does), an error answered by Tuanjie (such as a package too old to support `tloop status`), an unreadable answer, or an unknown option. These print an error to stderr, nothing to stdout, and exit 1.

## Output

Returns JSON with:

- `State`: one of the states above
- `Ready`: `true` only for `Ready`
- `Message`: what the state means for this Editor
- `NextActions`: what to do next; an empty array for `Ready`
- `RunningToolName`, `RunningToolElapsedSeconds`, `RunningToolPhase` (`Executing` or `WaitingForMainThread`): the running command; `Busy`, and `BlockedByDialog` while a command runs
- `IsPlaying`, `IsPaused`, `IsCompiling`, `IsUpdating`: Tuanjie's play and compile state; only when Tuanjie answered and has recorded that state
- `SecondsSinceLastMainThreadTick`: seconds since Tuanjie's main thread last ran; only when Tuanjie answered
- `UnityProcessRunning`: whether Tuanjie is running for this project; `ServerUnavailable` and `NotRunning` only, and `BlockedByDialog` found while the server did not answer
- `ConnectionError`: the connection error as reported; `NotResponding`, `ServerUnavailable`, `NotRunning`, and `Unreachable` only, and `BlockedByDialog` found while the server did not answer
- `Dialogs`: each open dialog's `Title`, `Message`, and `Buttons` (button labels; empty when tloop cannot read them); `BlockedByDialog` only
- `DialogCheckError`: why tloop could not look in Tuanjie's windows for a dialog; macOS and Windows only, with `Busy`, `MainThreadBlocked`, or `ServerUnavailable`

Play Mode does not change the state: a paused Play Mode session is `Ready`, with `IsPlaying` and `IsPaused` set.

## Notes

- Exits 0 only for Ready.
- When `Busy` does not clear and the running command shows no sign of finishing, stop it with `tloop cancel`; its outcomes, what it cannot cancel, and what to do after it are in [references/cancel.md](references/cancel.md).
- It never focuses the window, launches Tuanjie, or takes the command slot, so it is safe to poll.
- A sandboxed shell that cannot open the Unix socket gets an error, not NotRunning; run it outside the sandbox.
- tloop cannot detect Safe Mode itself; ServerUnavailable that does not clear is the signal. On macOS and Windows its prompt shows as BlockedByDialog; `tloop launch` presses Ignore.
- Before reporting MainThreadBlocked it wakes the Editor's main loop and asks again about a second later, so an Editor that is only idle in the background reads as Ready.
- Before sending any command other than `status`, `get-dialogs`, `cancel`, `press-dialog-button`, and `focus-window`, tloop asks the Editor once whether a modal dialog holds it; when one does, tloop does not send it and fails at once with `BLOCKED_BY_DIALOG` (`was not run`). A dialog it does not find there is caught as follows. A command sent while the main thread is blocked waits for it. When tloop finds a dialog while the command waits, the command fails with `BLOCKED_BY_DIALOG` (`was not run`) within seconds (about 3 on macOS); `compile` (and `run-tests` that compiles first) answers fast as `had already started`. On Tuanjie 1.9.1 Linux, and with an older package except for compile, the wait follows the rules below. Once the stall reaches 30 seconds before any of the command's code has run, Tuanjie stops the wait and the command fails with `MAIN_THREAD_BLOCKED` (nothing ran, so running it again after the block clears is safe). When tloop finds a dialog behind the block (the platform note below says where it can look), it fails with `BLOCKED_BY_DIALOG` instead, otherwise with `MAIN_THREAD_BLOCKED`; when that Message says the command `was not run`, nothing ran either, so follow its `NextActions`. A command whose own code blocks the main thread is not stopped; the CLI prints a stall notice on stderr every 60 seconds instead and gives up after 5 minutes with `UNITY_EDITOR_UNRESPONSIVE`. When a dialog holds the main thread after the command started, tloop stops waiting with `BLOCKED_BY_DIALOG` whose Message says the command `had already started`; do not rerun it blindly; follow its `NextActions`, whose last step says whether running it again is the way on (as for `tloop compile`, which reattaches). `execute-dynamic-code` is the exception: its own snippet may block the main thread for as long as it runs.
- For `BlockedByDialog`, read `Title`, `Message`, and `Buttons`, and press the button that fits what the user asked for with `tloop press-dialog-button --label "<button>"`. Ask the user first when no button clearly fits or the choice would discard work (such as `Don't Save`). When `NextActions` says tloop cannot press the dialog, show it to the user and ask them to answer it in Tuanjie. Pressing, its errors, and what it can press are in [references/dialog.md](references/dialog.md).
- Tuanjie 1.9.1 has no Unity 6000.7 dialog events. tloop uses operating-system window inspection for dialogs and startup prompts: on Windows as soon as an answer that would say `Ready`, `Compiling`, or `ImportingAssets` shows a main thread that has not run for 0.3 seconds; on macOS when the main thread has not run for 1 second or more and still has not when asked again; on both when it has not run for 5 seconds or more while a command runs, or when Tuanjie runs without its server. On Linux it does not look: on Tuanjie 1.9.1 such a dialog shows as `Busy`, `MainThreadBlocked`, or `ServerUnavailable`.
- On macOS that look needs Accessibility and Automation (System Events) access for the app that runs tloop, and the first one can make macOS ask whether the app may control System Events; on Windows it needs no permission. When it fails, the state is unchanged and, once the main thread has not run for 5 seconds or more, `DialogCheckError` says why; a dialog may still be open.
