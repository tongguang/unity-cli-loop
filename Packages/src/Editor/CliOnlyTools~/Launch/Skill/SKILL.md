---
name: tloop-launch
description: "Launch or restart Tuanjie Editor. Use only when the Editor is not running or stays frozen after retries — not as a health check after a failed command; it brings the Editor window to the foreground as a side effect."
---

# tloop launch

Launch Tuanjie Editor with the correct version for a project.

`tloop launch` is not fire-and-forget. When Tuanjie needs to start or restart, the command waits
until Tuanjie is actually ready for CLI operations before it exits.

## Usage

```bash
tloop launch [project-path] [options]
```

## Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `project-path` | string | Optional. Use only when the target Tuanjie project is not in the current directory. |
| `-r, --restart` | flag | Quit running Tuanjie and restart |
| `-q, --quit` | flag | Quit an existing Tuanjie process for the project without launching |
| `--editor-version <version>` | string | Use this Tuanjie Editor version instead of ProjectVersion.txt |
| `--editor-path <path>` | string | Use an explicit Editor executable instead of automatic discovery; cannot be combined with --editor-version |
| `-p, --platform <P>` | string | Build target (e.g., StandaloneOSX, Android, iOS) |
| `--max-depth <N>` | number | Search depth when project-path is omitted (default: 3, -1 for unlimited) |

## Output

May print status/progress lines before the final JSON payload, such as project path, detected Tuanjie version, or readiness wait messages.

The final JSON payload includes:

- `Success`: whether the command completed
- `Ready`: whether Tuanjie CLI Loop is ready for commands
- `ServerReady`: whether the Tuanjie CLI Loop server accepted requests
- `ProjectIpcReady`: whether the project IPC path accepted tool requests
- `AlreadyRunning`: whether an existing Tuanjie process was reused
- `Launched`: whether this command launched a Tuanjie process
- `Restarted`: whether this command stopped an existing process and launched a new one
- `Quit`: whether this command stopped Tuanjie without launching a new process
- `PreviousProcessId`: process ID stopped by restart or quit, when available
- `StopMethod`: how restart or quit stopped that process, omitted when there was none — `Quit` when Tuanjie exited on request, `ForceKill` when it had to be killed
- `CurrentProcessId`: current Tuanjie process ID, when available
- `ProjectRoot`: resolved project root
- `Message`: readiness summary

## When not to use

A single failed, cancelled, or busy command (e.g. right after a domain reload) is not a reason
to launch — retry the command instead. Running launch while the Editor is up brings the Tuanjie
window to the foreground, which disrupts the user. Reach for launch only when Tuanjie is not
running, or still does not respond after retries.

## Notes

- Automatically finds the Tuanjie version recorded in `ProjectSettings/ProjectVersion.txt`, checking Tuanjie Hub's registered installs and default directories
- Custom installs registered in Hub are discovered automatically; use `--editor-path` only for an executable outside those locations
- If Tuanjie is already running, focuses the existing window and verifies tool readiness
- `--restart` and `--quit` ask Tuanjie to quit first: it saves its window layout and discards unsaved scene changes without a save dialog. Tuanjie is force-killed instead, losing the layout, when it cannot take the request (frozen, or a package that predates it) or does not exit within 30 seconds
- If process scan is blocked (e.g. sandboxed `ps`), plain launch falls back to IPC probing; `--restart` and `--quit` still fail because they need the process id
- New Tuanjie processes ignore project compiler errors during Editor startup by default when run by dispatcher 3.0.1-beta.8 or newer; run `tloop update` first if an older dispatcher is installed
- `--editor-version` and `--editor-path` only affect new processes; use either with `--restart` to replace an already-running Editor
- The command waits until Tuanjie finishes startup and the CLI can connect before returning
