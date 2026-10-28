---
name: tloop-get-project-info
description: "Get Tuanjie project information via tloop CLI. Use when you need to check project settings, Tuanjie version, platform, or other project metadata."
internal: true
---

# tloop get-project-info

Get detailed Tuanjie project information.

## Usage

```bash
tloop get-project-info
```

## Parameters

None.

## Output

Returns JSON with:

- `ProjectName`: Application product name
- `CompanyName`: Company name
- `Version`: Application version
- `UnityVersion`: Tuanjie Editor version
- `Platform`: Current platform
- `DataPath`: Assets folder path
- `IsEditor`: Whether running in editor
- `IsPlaying`: Whether in play mode
- `DeviceType`: Device type
- `OperatingSystem`: OS information
- `ProcessorType`: CPU type
- `SystemMemorySize`: RAM size in MB
- `GraphicsDeviceName`: GPU name

## Notes

This is a sample custom tool demonstrating how to create Tuanjie CLI Loop tools.
