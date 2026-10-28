---
name: tloop-get-version
description: "Get Tuanjie and project information via tloop CLI. Use when you need to verify Tuanjie version, check project settings (ProductName, CompanyName, Version), or troubleshoot environment issues."
internal: true
---

# tloop get-version

Get Tuanjie version and project information.

## Usage

```bash
tloop get-version
```

## Parameters

None.

## Output

Returns JSON with:

- `UnityVersion`: Tuanjie Editor version
- `Platform`: Current platform
- `DataPath`: Assets folder path
- `PersistentDataPath`: Persistent data path
- `TemporaryCachePath`: Temporary cache path
- `IsEditor`: Whether running in editor
- `ProductName`: Application product name
- `CompanyName`: Company name
- `Version`: Application version
- `Ver`: Tuanjie CLI Loop package version
