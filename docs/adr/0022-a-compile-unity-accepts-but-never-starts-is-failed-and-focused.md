# A Compile the Editor Accepts but Never Starts Is Failed, and Focused First

Date: 2026-11-08

## Decision

A compile request that Tuanjie has accepted (`EditorApplication.isCompiling` is true) but has not
started (`CompilationPipeline.compilationStarted` has not fired) is a distinct state, "pending
start", and both sides act on it from Tuanjie's own evidence:

- Tuanjie counts the pending time in main-thread ticks and publishes it on `get-compile-status` as
  the additive field `CompileStartPendingSeconds`. A blocked main thread never counts: a compile
  the Editor has not reached is not a compile the Editor is withholding.
- Once the pending time reaches 10 seconds, the CLI tries to bring the Editor to the front once
  per wait and prints one stderr line saying why. When possible, it restores the previous window
  at the end of the wait. Only an answered status poll can trigger the focus attempt; a silent
  Editor (#3228) never triggers it. A package that predates the field omits it, which the CLI reads
  as zero.
- Once the pending time reaches 60 seconds, Tuanjie fails the CLI request with `COMPILE_NOT_STARTED`.
  Nothing is cancelled inside Tuanjie: the compile stays queued there and runs once the Editor lets
  it, and only the CLI's wait ends. The result's `Details` carry what Tuanjie knows
  (`CanReloadAssemblies`, `EditorWindowActive`, `HotControl`, the Editor flags) and what the CLI knows
  (`CliFocusAttempted`,
  `CliFocusSucceeded`, `HeldMouseButtons` from the OS on Windows). `NextActions` say to click inside
  the Tuanjie window and rerun. Because the compile stays queued inside Tuanjie, `get-compile-status`
  reports a stored `COMPILE_NOT_STARTED` result as ready even while `isCompiling` is still true.
- Both thresholds apply on every platform, because the condition they react to is Tuanjie's, not
  the OS's. The 60-second failure therefore happens everywhere. The 10-second focus is attempted
  everywhere too, but it can only succeed where `unityprocess` can focus a window (Windows and
  macOS); elsewhere the attempt fails, is logged, and the wait goes on to the 60-second failure.

## Context

[Upstream issue #3434](https://github.com/hatayama/unity-cli-loop/issues/3434): on Windows,
`uloop compile` (also the implicit compile of `uloop run-tests`) sometimes
stalled before compilation started while the Editor was in the background, and clicking inside the
Editor made it start. The reporter uses an input-sharing tool. The stall was reproduced on a
Windows machine by swallowing a mouse-up event while the Editor was in the background: Unity then
keeps `isCompiling` true but withholds the compile for as long as its internal
`CanReloadAssemblies()` reports a mouse button held, and `SetForegroundWindow` on the Editor
clears it at once. Resetting `GUIUtility.hotControl` does not.

Before this decision the watchdog treated `isCompiling` alone as "started", so the stall looked
like an ordinary long compile until the CLI's 10-minute wait ran out, and nothing in the result
said what had happened. The ten-second focus that #2341 added for the same symptom had been removed
by ADR 0012, which reasoned from macOS: with the activity held, a compile that has not started is a
compile the Editor has not reached yet. That reasoning holds for App Nap but not for this stall,
where the Editor is ticking and has reached the request.

## Considered Options

- Focus on wall-clock time since the request, as #2341 did: rejected. It cannot tell a withheld
  compile from a main thread that is busy with something else (a large import, a dialog), which is
  what ADR 0012 removed it for.
- Focus from the Unity side (`EditorWindow.Focus`, `SetForegroundWindow` from the package):
  rejected. Unity cannot bring itself to the front on Windows without the foreground-lock dance the
  CLI already has in `unityprocess`, and the CLI is the process the user ran, so it is the one
  allowed to ask for the foreground.
- Clear the stuck state directly (reset `hotControl`, send a synthetic mouse-up): rejected.
  `hotControl` is not the flag Unity reads, and synthesizing input into another process is the
  behavior that caused the stall.
- Only fail, never focus: rejected. The report asked for either; failing alone leaves the user to
  click, and the focus is what makes the compile start without them.
- Only focus, never fail: rejected. When the focus does not help (no window, a remote session, a
  button physically held) the wait would still run to the 10-minute timeout with no explanation.
- Lower thresholds (5 s focus, 30 s abort): rejected. Ordinary compiles on large projects show a
  few seconds between request and `compilationStarted`, and a window focus that fires on them
  would move the user's windows on every compile.

## Consequences

- A compile that stays pending for 10 seconds makes `tloop` try to bring the Editor to the front.
  When supported, the previous window regains focus after the command finishes. ADR 0012's note that
  `compile` no longer focuses is superseded for this one condition and points here.
- `COMPILE_NOT_STARTED` is an additive `ErrorCode` and `CompileStartPendingSeconds` an additive
  field, so the protocol version does not change. An old CLI with a new package sees the error
  without the CLI-side `Details`; a new CLI with an old package never focuses.
- The result after an abort carries diagnostics the next report can be read from, instead of a
  timeout with nothing in it.
- A compile that is legitimately slow to start (over 60 seconds between request and
  `compilationStarted`) is now failed and must be run again; nothing seen so far comes close.

## Revisit When

- (a) Real compiles take more than 10 seconds to reach `compilationStarted` on some project and the
  focus moves windows on ordinary runs: raise `compileStartPendingFocusThreshold`.
- (b) Compiles that are legitimately slow to start hit the 60-second abort: raise
  `UnityCliLoopConstants.COMPILE_START_PENDING_TIMEOUT_MS`.
- (c) Unity stops withholding compiles on a believed-held mouse button, or exposes
  `CanReloadAssemblies` publicly: the reflection read in `CompileStartPendingDiagnostics` can go.
