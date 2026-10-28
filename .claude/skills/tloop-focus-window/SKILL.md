---
name: tloop-focus-window
description: "Bring the Tuanjie Editor window to front. Use when Tuanjie must be visible for visual checks or user-facing interaction."
---

# tloop focus-window

Bring Tuanjie Editor window to front using OS-level commands.

## Usage

```bash
tloop focus-window
```

## Output

Returns JSON with:

- `Success`: Whether the focus operation succeeded
- `Message`: Status message (e.g. `Tuanjie Editor window focused (PID: 12345)`, or the failure reason such as `Tuanjie project not found` / `No running Tuanjie process found for this project` / `Failed to focus Tuanjie window: <reason>`)

## Notes

- Useful before `tloop screenshot` to ensure the target window is visible
