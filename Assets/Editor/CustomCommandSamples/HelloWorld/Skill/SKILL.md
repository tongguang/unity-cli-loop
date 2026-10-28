---
name: tloop-hello-world
description: "Sample tloop hello-world tool. Use to verify custom tool wiring or inspect a minimal tool implementation example."
---

# tloop hello-world

Personalized hello world tool with multi-language support.

## Usage

```bash
tloop hello-world [options]
```

## Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--name` | string | `World` | Name to greet |
| `--language` | string | `english` | Language for greeting: `english`, `japanese`, `spanish`, `french` |
| `--no-include-timestamp` | flag | - | Exclude timestamp from response |

## Examples

```bash
# Default greeting
tloop hello-world

# Greet with custom name
tloop hello-world --name "Alice"

# Japanese greeting
tloop hello-world --name "太郎" --language japanese

# Spanish greeting without timestamp
tloop hello-world --name "Carlos" --language spanish --no-include-timestamp
```

## Output

Returns JSON with:

- `Message`: The greeting message
- `Language`: Language used for greeting
- `Timestamp`: Current timestamp (if enabled)

## Notes

This is a sample custom tool demonstrating:

- Type-safe parameter handling with Schema
- Enum parameters for language selection
- Value-less flag parameters
- Multi-language support
