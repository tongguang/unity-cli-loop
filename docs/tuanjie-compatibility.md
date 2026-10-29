# Tuanjie CLI Loop (`tloop`)

This branch targets Tuanjie only. The CLI discovers the Editor automatically from
`ProjectSettings/ProjectVersion.txt`, using a version such as `2022.3.62t15`.
It does not launch a Unity Editor or substitute a different release when the
requested Tuanjie version is missing.

The command is `tloop`. It installs alongside the original `uloop` used by Unity
projects. This branch does not install an `uloop` alias, read its environment
variables, reuse its cache, or migrate its installed skills.

| Surface | Tuanjie identity |
| --- | --- |
| Dispatcher / project runner | `tloop` / `tloop-project-runner` |
| Windows install directory | `%LOCALAPPDATA%\Programs\tloop\bin` |
| macOS / Linux command | `~/.local/bin/tloop` |
| Environment variables | `TLOOP_*` |
| Download cache | `tloop` under the platform cache directory |
| Project state / outputs / pin | `.tloop/` |
| Offline compiler state | `Library/tloop/` |
| Windows IPC | `\\.\pipe\tloop-TuanjieCliLoop-<project hash>` |
| macOS / Linux IPC | `/tmp/tloop-<uid>/TuanjieCliLoop-<project hash>.sock` |
| Installed skills | `tloop-*` |
| Optional skill group directory | `tuanjie-cli-loop/` |
| UPM package | `io.github.tongguang.tloop` |
| Releases and update source | `tongguang/unity-cli-loop` |

There is no TCP listener or starting port to configure. Both transports derive
their endpoint from the canonical project path, with a separate Tuanjie prefix.
Installing, updating, or uninstalling `tloop` must leave the original `uloop`
command, PATH block, caches, skills, and project state intact.

Skill names, descriptions, command examples, references, and the generated tool
catalog come from the package sources. The package's Editor installer and the CLI
use those same sources. `tloop skills install --global` installs the Tuanjie skills
with their own names; it does not replace `uloop-*` skills.

Builds write `dist/<platform>/tloop` and `tloop-project-runner` (with `.exe` on
Windows). Use these local binaries until this fork has published its own release.
Release automation owns the versions and signed pin manifests; existing upstream
archive hashes are not relabeled by hand. A missing Tuanjie release is an error,
not permission to download the original Unity CLI.

## GitHub packaging and first release

Native CLI packaging does not require an installed Editor. The existing Go build
jobs produce `tloop-dispatcher-<platform>` archives containing `tloop`, and
`tloop-project-runner-<platform>` archives containing `tloop-project-runner`.
Windows archives contain the corresponding `.exe` files. Installer assets,
checksums, release asset verification, and attestation policy use the fork's
repository identity. Project runner tags use `tloop-project-runner-v<version>`;
dispatcher tags remain `dispatcher-v<version>` within this repository.

The package release creates an ordinary UPM tarball from the complete package
tree without launching Unity or Tuanjie or using Unity Cloud signing. Its asset is
`io.github.tongguang.tloop-<version>.tgz`, including the source skills. Existing
workflow filenames and the `unity-package` release component are automation
identifiers and do not select an Editor.

Use the explicit [Release workflow](releasing.md) to publish native assets and
stamp their verified hashes into a later package commit. It uses the ordinary
workflow token and requires no GitHub App, approval environment, or cloud secrets.
Do not reuse upstream archives or manually relabel their pin entries.

Homebrew and winget publication are not configured. This branch does not publish
to OpenUPM; `tloop package install` uses the fork's Git URL, following `main` by
default or a release tag with `--version`.

No CI job starts a Unity or Tuanjie Editor. The inherited GameCI jobs tested
Unity and ran only when a `UNITY_LICENSE` secret was set, which this fork never
had, so they were removed. The unified `CI` workflow retains the five Editor-free
.NET unit test projects. A Tuanjie CI job would need a machine with an
installed and activated Tuanjie Editor, which GitHub-hosted runners do not
provide. Use the [local Tuanjie check](#local-tuanjie-check) instead.

## Editor discovery and local validation

Tuanjie also writes a product version such as `m_TuanjieEditorVersion: 1.10.3`.
Discovery uses `m_EditorVersion` (`2022.3.62t15` in that example), which matches the
installed Editor directory. Use that value when passing `--editor-version`.

Discovery checks `TuanjieHub/editors-v2.json` under the user's configuration
directory, then Tuanjie's default Hub install directories. Manually located Editors
registered in Hub work without additional options. Both `launch` and `compile-check`
use this discovery.

```sh
tloop launch
tloop compile-check
```

For installations outside these locations, pass the executable explicitly:

```powershell
tloop launch --editor-path 'D:\Engines\Tuanjie\Editor\Tuanjie.exe'
tloop compile-check --editor-path 'D:\Engines\Tuanjie\Editor\Tuanjie.exe'
```

`--editor-version` selects a different release through automatic discovery.
`--editor-path` selects an executable directly and bypasses project version lookup.
The two options are mutually exclusive. Replacing an already running Editor requires
`launch --restart`.

Process discovery recognizes `Tuanjie.exe` on Windows, `Tuanjie.app` on macOS,
and the `Tuanjie` executable on Linux. Batch Editors and asset import workers remain
excluded. Restart, quit, window focus, and
offline compile safety checks share this process discovery.

Crash diagnostics search Tuanjie's default Editor logs and check that a log
belongs to the requested project before using it. Tuanjie uses
`%LOCALAPPDATA%\Tuanjie\Editor\Editor.log` on Windows,
`~/Library/Logs/Tuanjie/Editor.log` on macOS, and
`~/.config/tuanjie/Editor.log` on Linux, as documented in the
[Tuanjie manual](https://docs.unity.cn/cn/tuanjiemanual/Manual/LogFiles.html).

Assembly and extension API names retain their existing spellings. The internal
JSON-RPC metadata key `uloop` also stays unchanged, so this branding change does
not require a new protocol version. Tuanjie still exposes `UnityEngine` and
`UnityEditor` APIs. Bundled compiler discovery already probes the Editor's actual
`Data` or `Contents` layout instead of depending on the executable's name.

The project's recorded Editor version must name a Tuanjie installation. A project
last saved by Unity must first be opened and saved in Tuanjie; `--editor-version`
or `--editor-path` can also select a Tuanjie Editor explicitly.

This repository's own development project uses Tuanjie 1.9.1 (`2022.3.62t9`).
Opening it in that Editor raised Input System to the Editor-bundled `1.14.4-t1`
and the Universal Render Pipeline to the built-in `14.2.0-t1`, and switched the
default registry to `packages.tuanjie.cn`.

These discovery changes do not establish compatibility with every Tuanjie release.
Tools using private Editor APIs, the bundled compiler, or hot reload still need
validation against the intended Editor version. Do not rewrite Unity version defines
or assembly names solely because they contain `Unity`.

Local validation on Windows imported this package into a separate project with
Tuanjie 1.10.3 (`2022.3.62t15`) and loaded all 50 package assemblies successfully.
With the headless Editor running, `status` reported `Ready`, `get-hierarchy`
succeeded, and `get-logs --log-type Error` returned no errors. `compile-check --all`
automatically found Tuanjie's bundled compiler and compiled 52 project assemblies
with zero errors and warnings.

After the `tloop` rename, the new Windows named pipe returned `Ready` and served a
hierarchy request. The old `uloop` binary could not connect to that same project,
confirming transport isolation. Tuanjie's batch Test Runner passed all 90 scoped
tests for endpoint contracts, CLI installers, pin synchronization, skill roots,
frontmatter parsing, line endings, and rejection of original `uloop-*` skill sources
and installed directories. The final `tloop compile-check --all` compiled 57
assemblies with zero errors and warnings, including the isolated test assemblies.
The CLI generated 22 skills for each of the repository's `.agents` and `.claude`
targets from the renamed sources.

In a headless Editor, the CLI compile wait currently excludes the batch process
from its process scan and can report `UNITY_NOT_RUNNING` across a domain reload
while the Editor is still alive. The scoped tests therefore used the batch Test
Runner directly. This does not establish graphical compile or startup behavior.

The existing POSIX manifest CR rejection test fails under this host's Git Bash
both before and after these changes. Native Windows pin-resolution tests passed;
the Bash text-mode difference remains outside this rename.

Graphical Editor startup did not reach CLI readiness during this session; a later
desktop session verified it (see below). The successful headless checks do not
establish that graphical startup completed.

That project used Input System 1.7.0: the Tuanjie package registry did not provide
the development project's Input System 1.14.2. Package versions in an existing
project still need to be available from its configured registry.

With Tuanjie 1.9.1 the development project opened from Tuanjie Hub reached
`Ready` in the graphical Editor, and the released dispatcher `3.14.1` compiled it
with zero errors and warnings. A scoped EditMode run of the third-party migration,
pause-point, transform-worker, and hot-reload package-source tests passed 957 of
958 tests with one skipped and none failed. The pause-point tests need Debug code
optimization; in an Editor running in Release, run
`tloop set-code-optimization debug` before them.

A later graphical-Editor session with Tuanjie 1.9.1 on Windows exercised the
runtime tools with a dispatcher built from this checkout. `launch` started the
Editor and returned once it was ready; run again, it reported `AlreadyRunning`.
`launch -r` asked the Editor to quit (`StopMethod: Quit`) and was ready again
after 49 seconds. `compile` returned zero errors and warnings after the domain
reload that its C# changes caused. `control-play-mode` started and stopped Play
Mode. In Play Mode, `screenshot --capture-mode rendering` captured the 1080x1920
Game View, and `record-video --seconds 3` recorded 96 frames of the
HotReloadSpinningImage harness, 95 of them different from the previous frame.
`scripts/regression-harness-hot-reload.sh` passed: the patched method body
changed the logged marker, and `--revert-all` restored it. On the
KeyStateAfterPauseInterruption harness, a file:line
`enable-pause-point --await --trigger "simulate-keyboard --action Press --key Space --duration 5"`
hit after 1.9 seconds and captured `isHeld = True`, with Debug code
optimization. `scripts/check-tuanjie.ps1` passed.

The development project enters Play Mode with Domain Reload disabled, so that
session did not exercise a Play-entry domain reload. `focus-window` failed
because Windows refused the foreground change for a caller that was not in the
foreground itself (foreground lock); click the Editor window or its taskbar icon
instead. When `launch` runs from a shell that ends every child process as a
command returns, the Editor it started exits with that command. That comes from
the calling environment: an Editor launched from a detached job kept running.

The historical release-please config uses the upstream import commit as
`bootstrap-sha`, so its first release notes contain this fork's changes rather
than replaying the full upstream history. Version fields are still generated by
release-please. Current explicit releases are described in [Releasing](releasing.md).

The first published fork-native releases are dispatcher `3.14.0` and project
runner `3.13.0`. The package requires dispatcher `3.14.0`, because earlier
dispatcher versions belong to the upstream CLI and are not published under this
fork's identity. The pin's archive digests were generated from the published
dispatcher's verified attestations. The public Windows dispatcher archive passed
the product's own offline attestation verifier and returned `3.14.0` when run
from an isolated extraction directory; both tagged installer scripts were reachable.

On this Windows host, GitHub CLI 2.97.0 reported `verifying with issuer
"sigstore.dev"` without details when checking the public archive. The product's
embedded verifier succeeded against that same downloaded bundle and tag commit;
the publisher's GitHub-hosted Linux verification also succeeded.

## Local Tuanjie check

`scripts/check-tuanjie.ps1` checks this project against the Tuanjie Editor that
already has it open. It confirms the Editor can take a command, compiles the
project and fails on any error or warning, then runs the `CliVersionComparerTests`
EditMode tests. It exits 1 at the first failing step. Run it after merging
upstream changes and before starting an explicit release.

```powershell
pwsh -NoProfile -File ./scripts/check-tuanjie.ps1
pwsh -NoProfile -File ./scripts/check-tuanjie.ps1 -TestFilterType regex -TestFilterValue CliPin
pwsh -NoProfile -File ./scripts/check-tuanjie.ps1 -SkipTests
```

The default tests only show that the Test Runner works under Tuanjie. To cover a
change, pass `-TestFilterType` and `-TestFilterValue` with the values
`tloop run-tests` accepts, except `all`: the unfiltered suite takes far too long
for this check.

The script uses this checkout's `dist` binary; rebuild it with
`scripts/build-go-cli.sh` after CLI source changes, or pass `-TloopPath`. It never
launches, restarts, or focuses the Editor. It stops when the Editor is in Play
Mode, because a compile would discard the Play session, and it runs the tests
with `--unsaved-changes fail`, so unsaved Scene changes stop the test run instead
of being saved. `scripts/test-check-tuanjie.ps1` covers these outcomes with a
fake `tloop`.
