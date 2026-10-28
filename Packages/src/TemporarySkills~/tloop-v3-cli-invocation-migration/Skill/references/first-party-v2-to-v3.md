# First-Party V2 to V3 CLI Migration

Use this reference as the canonical option migration map. Search results are
only candidates. Edit a match only after the surrounding context proves it is
a V2 `tloop` invocation.

This migration is option-only by default. Do not reformat surrounding text or
scripts, and do not change non-`tloop` commands. A valid edit only replaces,
adds, or removes a `tloop` option token and the boolean value attached to that
option in the same invocation.

Markdownlint is not required validation for this migration. If Markdownlint
runs and reports existing or unrelated formatting issues, ignore that result,
continue the option migration, and report the lint result as non-blocking. Do
not fix Markdownlint findings unless the migration itself created them.

## Search Checklist

Prefer `rg` when available, but any repository search tool is acceptable.

- Search `tloop` first and inspect command examples, shell scripts, PowerShell scripts, and agent skills.
- Search boolean-looking CLI syntax: `--` plus nearby `true` or `false`, including `--flag true`, `--flag=false`, and inline Markdown command examples.
- Search renamed first-party option names: `wait-for-domain-reload`, `reload-external-scene-changes`, `force-recompile`, `save-before-run`, `show-overlay`, `include-components`, `include-inactive`, and `compile-only`.
- Treat renamed-option hits as command-scoped. `wait-for-domain-reload` and
  `include-inactive` also exist as valid V3 positive flags on other
  commands; check Same-Name Options on Other Commands before editing them.
- Search removed or renamed first-party commands only to report them as
  out-of-scope command migration candidates: `get-project-info`, `get-version`, `record-input`,
  `unity-search`, `execute-menu-item`, `get-menu-items`,
  `get-unity-search-providers`, `get-provider-details`, and `capture-window`.
- Skip generated installed skill copies under `.agents`, `.claude`, `.codex`, `.cursor`, `.gemini`, `.windsurf`, `.agent`, or equivalent target folders unless the user explicitly asks to migrate installed copies.

## Boolean Argument Rules

| V2 form | V3 form |
| --- | --- |
| `--flag true` | `--flag` when the V3 option is a positive default-false boolean |
| `--flag=true` | `--flag` when the V3 option is a positive default-false boolean |
| `--flag=false` | remove the option when the V3 default is already false |
| `--flag false` | remove the option when the V3 default is already false |
| `--flag true` | remove the option when the V3 default is already true |
| `--flag=true` | remove the option when the V3 default is already true |
| `--flag false` | use the V3 negative option when the V3 default is true |
| `--flag=false` | use the V3 negative option when the V3 default is true |

For first-party boolean options that are not listed in Special First-Party
Options, run `tloop <command> --help`. Every V3 flag is printed with its
default: `default: enabled` means the V3 default is true, and
`default: disabled` means it is false.

For third-party tools, inspect the current tool schema or docs before choosing the replacement. Do not infer third-party negative flags from first-party conventions.

## Special First-Party Options

| V2 command | V2 option | V3 replacement |
| --- | --- | --- |
| `tloop compile` | `--force-recompile true` | `--force-recompile` |
| `tloop compile` | `--force-recompile false` | remove |
| `tloop compile` | `--wait-for-domain-reload true` or bare `--wait-for-domain-reload` | remove |
| `tloop compile` | `--wait-for-domain-reload false` | `--no-wait-for-domain-reload` |
| `tloop compile` | `--reload-external-scene-changes true` | remove |
| `tloop compile` | `--reload-external-scene-changes false` | `--stop-on-external-scene-changes` |
| `tloop run-tests` | `--save-before-run true` or bare `--save-before-run` | remove |
| `tloop run-tests` | `--save-before-run false` | `--unsaved-changes fail` |
| `tloop replay-input` | `--show-overlay true` | remove |
| `tloop replay-input` | `--show-overlay false` | `--no-show-overlay` |
| `tloop get-hierarchy` | `--include-components true` | remove |
| `tloop get-hierarchy` | `--include-components false` | `--no-include-components` |
| `tloop get-hierarchy` | `--include-inactive true` | remove |
| `tloop get-hierarchy` | `--include-inactive false` | `--no-include-inactive` |
| `tloop execute-dynamic-code` | `--compile-only true` | `--compile-only` |
| `tloop execute-dynamic-code` | `--compile-only false` | remove |

## Same-Name Options on Other Commands

These option names appear in the table above for one command but are valid
V3 syntax on another command. Never migrate them by token match alone. On the
other command, only the already-bare flag is exempt: value-bearing forms such
as `--flag true` or `--flag=false` still follow the Boolean Argument Rules
against that command's default.

| Option | Migrate on | Also valid V3 on |
| --- | --- | --- |
| `wait-for-domain-reload` | `tloop compile` (see Special First-Party Options) | `tloop execute-dynamic-code`, where bare `--wait-for-domain-reload` is a valid V3 default-false flag |
| `include-inactive` | `tloop get-hierarchy` (see Special First-Party Options) | `tloop find-game-objects`, where bare `--include-inactive` is a valid V3 default-false flag |

Bare `--force-recompile` on `tloop compile` and bare `--compile-only` on
`tloop execute-dynamic-code` are already valid V3 syntax and need no edit.

## Removed First-Party Commands

| V2 command | V3 handling |
| --- | --- |
| `tloop get-project-info` | Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
| `tloop get-version` | Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
| `tloop record-input` | Removed. Create recordings in the Tuanjie Editor from **Window > Tuanjie CLI Loop > Recordings**. Report only unless the user explicitly asks for removed command migration. |
| `tloop unity-search` | Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
| `tloop execute-menu-item` | Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
| `tloop get-menu-items` | Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
| `tloop get-unity-search-providers` | Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
| `tloop get-provider-details` | Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
| `tloop capture-window` | Renamed to `tloop screenshot`. Report only unless the user explicitly asks for removed command migration. Do not guess from the command name alone. |
