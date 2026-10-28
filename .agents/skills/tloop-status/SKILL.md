---
name: tloop-status
description: "Check whether the Tuanjie Editor for this project can take a tloop command now, without side effects; also covers `tloop cancel` for stopping the command it is running. Use before sending commands when Tuanjie may be starting, compiling, busy, or closed, and when BUSY does not clear or a command does not finish."
---

# tloop status

Report whether the Tuanjie Editor for this project can take a tloop command now, without changing anything.

## Usage

```bash
tloop status
tloop status --project-path <path>
```

## States

The states are checked in this order, and the first one that matches is reported.

| State | Meaning | Exit code |
|-------|---------|-----------|
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
- `RunningToolName`, `RunningToolElapsedSeconds`, `RunningToolPhase` (`Executing` or `WaitingForMainThread`): the running command; `Busy` only
- `IsPlaying`, `IsPaused`, `IsCompiling`, `IsUpdating`: Tuanjie's play and compile state; only when Tuanjie answered and has recorded that state
- `SecondsSinceLastMainThreadTick`: seconds since Tuanjie's main thread last ran; only when Tuanjie answered
- `UnityProcessRunning`: whether Tuanjie is running for this project; `ServerUnavailable` and `NotRunning` only
- `ConnectionError`: the connection error as reported; `NotResponding`, `ServerUnavailable`, `NotRunning`, and `Unreachable` only

Play Mode does not change the state: a paused Play Mode session is `Ready`, with `IsPlaying` and `IsPaused` set.

## Notes

- Exits 0 only for Ready.
- When `Busy` does not clear and the running command shows no sign of finishing, stop it with `tloop cancel`; its outcomes, what it cannot cancel, and what to do after it are in [references/cancel.md](references/cancel.md).
- It never focuses the window, launches Tuanjie, or takes the command slot, so it is safe to poll.
- A sandboxed shell that cannot open the Unix socket gets an error, not NotRunning; run it outside the sandbox.
- tloop cannot detect Safe Mode directly; ServerUnavailable that does not clear is the signal.
- Before reporting MainThreadBlocked it wakes the Editor's main loop and asks again about a second later, so an Editor that is only idle in the background reads as Ready.
