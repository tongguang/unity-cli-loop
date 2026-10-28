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

Before publishing the fork's first release:

- Enable Actions and configure the `cli-release` environment in this repository.
  The package workflow needs no signing environment or cloud secrets.
- Configure the GitHub App described in `dispatcher-pin-release-order.md` for
  unattended pin pushes and package PR merges. Without the App, those steps skip;
  use the documented pin-stamping and package-merge helpers for this release.
- Use the existing release flow to publish the fork's native assets and stamp
  their verified hashes into the package pin before publishing the package.
  Do not reuse upstream archives or manually relabel their pin entries.
- Configure Homebrew, winget, and OpenUPM separately if those channels are wanted.
  The source configuration alone does not register or publish the renamed
  package in any external registry.

The inherited `unity-compile-check-and-test-runner.yml` and
`unity-editmode-tests.yml` still run Unity Editors, including Unity 6000 releases.
They have not been migrated to Tuanjie and are not evidence of Tuanjie
compatibility. A Tuanjie CI job needs an available Tuanjie Editor and its own
working activation setup; changing the matrix version string alone does not
provide either. The first release check found no repository secrets or GitHub
App configuration.
The package workflow was then changed to ordinary UPM packing at the user's request.

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
default registry to `packages.tuanjie.cn`. The EditMode workflow still tests with
Unity: its GameCI jobs run the Unity versions named in their matrix, and its
Release leg downloads the Editor named in `ProjectVersion.txt` from Unity's
servers, which serve no Tuanjie build. These jobs run only when the
`UNITY_LICENSE` secret is set, and the Release leg cannot start against this
project until it installs Tuanjie instead.

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

Graphical Editor startup did not reach CLI readiness during this session and still
needs verification in a desktop session. The successful headless checks do not
establish that graphical startup completed.

That project used Input System 1.7.0: the Tuanjie package registry did not provide
the development project's Input System 1.14.2. Package versions in an existing
project still need to be available from its configured registry.

With Tuanjie 1.9.1 the development project opened from Tuanjie Hub reached
`Ready` in the graphical Editor, and the released dispatcher `3.14.1` compiled it
with zero errors and warnings. A scoped EditMode run of the third-party migration,
pause-point, transform-worker, and hot-reload package-source tests passed 957 of
958 tests with one skipped and none failed. The pause-point tests need Debug code
optimization, which the CI workflow sets with `-debugCodeOptimization`; in an
Editor running in Release, run `tloop set-code-optimization debug` before them.

The fork's release-please config uses the upstream import commit as
`bootstrap-sha`, so its first release notes contain this fork's changes rather
than replaying the full upstream history. Version fields are still generated by
release-please. See the [release-please manifest documentation](https://github.com/googleapis/release-please/blob/main/docs/manifest-releaser.md).

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
