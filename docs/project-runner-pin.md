# Project runner pin

`Packages/src/project-runner-pin.json` (mirrored byte-identically to
`.tloop/project-runner-pin.json` by `CliPinSynchronizer`) is the single source
for cross-component version requirements. Its required fields:

- `projectRunnerVersion` — the project runner release the dispatcher must run for this package.
  Stamped by the explicit Release workflow; never edit by hand.
- `minimumDispatcherVersion` — the semver floor the package requires of the globally installed
  dispatcher. The dispatcher force-updates itself when it is older than this value, and the
  package reads it (via `CliPinReader`) for setup and installation checks. This is the only
  manually maintained minimum-version declaration; raise it only when the package genuinely
  needs a newly published dispatcher, not because the dispatcher implementation changed.
- `dispatcherReleaseTag` and `dispatcherArchiveManifest` — the provenance-pinned dispatcher
  release used for first installation and its verified asset hashes. Stamped by automation
  against a published release: after native publication, the Release workflow verifies
  its attestation and writes a later package-version commit containing the stamp.
  The pin is checked during CI and package publication. Never edit it by hand —
  `VerifyDispatcherPinSubjects` requires the
  manifest to match the published release's verified subjects exactly, so a hand-written value
  cannot pass CI (see `docs/dispatcher-pin-release-order.md`). Consumers: Unity's **Install CLI**
  path (via `CliPinReader` / `NativeCliCommandBuilder`) and the terminal installers
  (`scripts/install.sh` / `scripts/install.ps1`) when `TLOOP_ARCHIVE_MANIFEST` is unset — they
  fetch this pin and use `dispatcherReleaseTag` as the install target plus
  `dispatcherArchiveManifest` as the attestation digest list.

There is no dispatcher⇄package integer contract generation; the pin's semver
floor is the only dispatcher gate. The IPC `protocolVersion` pair (see
`docs/protocol-version.md`) is the only integer generation in the system.

## Local override: `TLOOP_PROJECT_RUNNER_PATH`

The environment variable `TLOOP_PROJECT_RUNNER_PATH` (defined in
`cli/dispatcher/internal/nativepath/path.go`) makes the dispatcher run the
project runner binary at that path instead of resolving one from the pin.
`resolveDispatcherRealCLI` checks it before everything else — pin validation,
the sibling binary next to the dispatcher, the version cache, and the GitHub
release download are all skipped. The path must point at an existing executable
file, or the dispatcher fails with an explicit error rather than falling back.

This override exists for dogfooding checkouts: release preparation stamps
`projectRunnerVersion` ahead of the matching GitHub release, so the normal
download path 404s until the release is published. Pointing the variable at a
locally built binary (e.g. `dist/darwin-arm64/tloop-project-runner`, refreshed
via `scripts/build-go-cli.sh`) lets you exercise unreleased project-runner and
`cli/common` changes against real Unity projects before merge. Unset the
variable to return to normal pin-resolved behavior.

Related overrides in the same file: `TLOOP_INSTALL_DIR` (dispatcher install
directory) and `TLOOP_CACHE_DIR` (project runner download cache).

## Which runner actually ran

`resolveDispatcherRealCLI` picks the runner in this order, and **no response field says which
branch won**:

1. `TLOOP_PROJECT_RUNNER_PATH` (env override).
2. The sibling `tloop-project-runner` next to the dispatcher binary — taken only when the pin's
   `projectRunnerVersion` equals the version compiled into *that* dispatcher
   (`clicontract.ProjectRunnerVersion()`).
3. `<cache root>/versions/<projectRunnerVersion>/<platform>/` (already downloaded).
4. GitHub release download.

Read this before trusting a dogfooding result. All four points were measured on 2026-07-27.

- Running `dist/<platform>/tloop` **without** the override still uses your local build, because
  step 2 finds `dist/<platform>/tloop-project-runner` beside it. Verified by pointing
  `TLOOP_CACHE_DIR` at an empty directory: the command succeeded and the directory stayed empty,
  so neither the cache nor a download was involved. The override is therefore not what makes a
  run local — the dispatcher you typed is.
- Running the installed `tloop` (`~/.local/bin/tloop`, which has no sibling runner) without the
  override silently uses whatever step 3 or 4 supplies, i.e. the released runner.
- Step 2 compares against a version baked into the dispatcher at build time, so it stops
  applying as soon as the pin moves ahead of your last `scripts/build-go-cli.sh`. That does not
  necessarily fail loudly: when the cache already holds that version, step 3 quietly succeeds
  with the released runner instead.
- A version number does not identify the binary. A local build and the released runner both
  answer `3.0.0-beta.58` to `tloop-project-runner --version` while their SHA-256 differ. (`version`
  as a *subcommand* is rejected by the runner on purpose — it belongs to the dispatcher — so only
  the flag form reports the runner's own version.)

Until the dispatcher reports the resolved runner path, comparing SHA-256 by hand is the only
reliable way to establish which binary served a verification run.

## Pin format discipline

The pin evolves additively only — never delete or rename an existing field.
The forced-update instruction (`minimumDispatcherVersion`) travels inside the
pin, so an old dispatcher that cannot parse a new pin never learns it must
update. For the same reason the dispatcher must stay lenient when reading pins
written by older packages.

## Periodic dispatcher self-update

Separately from the `minimumDispatcherVersion` floor, the dispatcher checks for the newest
release on its own, independent of any pin, Tuanjie project, or working directory.

- **When**: on every command, at most once per 24 hours. Excluded are `tloop update`,
  `install`, `uninstall`, `version`, and help and version requests (including a project-scoped
  `--version` that is forwarded to the runner). `install` and `uninstall` are excluded because
  installers and the Tuanjie settings window run them, and a second installer must not start
  from inside the first.
- **Effect**: a newer release is installed while the command runs, but that command finishes
  on the old dispatcher; the next command uses the new one.
- **Where it runs**: only when the running `tloop` sits in the install directory
  (`TLOOP_INSTALL_DIR`, otherwise the OS default). The installer only replaces that
  directory, so a dev build or a hand-extracted copy would rewrite the global install every
  day while staying old.
- **Never runs**: under Homebrew or winget, or with `TLOOP_DISABLE_SELF_UPDATE=1|true`. The
  installer child process always gets `TLOOP_DISABLE_SELF_UPDATE=1`.
- **Record and lock**: the cache root (`TLOOP_CACHE_DIR`, otherwise the OS cache directory)
  holds `dispatcher-update.json` (the last check, written *before* the lookup; no lookup runs
  when it cannot be written) and `dispatcher-update.lock`. The lock is an OS file lock on that
  file (`flock` on macOS / Linux, `LockFileEx` on Windows), so it is released when the holding
  process exits and never goes stale; the file itself stays in place. The required update for
  `minimumDispatcherVersion` takes the same lock; when another process holds it, the command
  fails with a retryable `CLI_UPDATE_REQUIRED` instead of starting a second installer.
  `tloop update` takes it before touching the hold and exits with a retryable error while
  another process holds it. When the periodic update already installed a release that meets
  the floor in the same process, the required check asks for a retry instead of installing
  again; a periodic install below the floor is followed by the required update.
- **Required update**: below the `minimumDispatcherVersion` floor, the dispatcher installs the
  newest release on its channel, or the floor itself when that release is older (a beta floor
  while the newest stable release is older installs the beta, and later automatic updates then
  follow the beta channel). If the dispatcher is still below the floor afterwards, the command
  fails with a manual-update error instead of asking for a retry that would repeat forever.
  A dispatcher that runs from outside the install directory does not update itself even below
  the floor: the installer would replace a different copy, so the command asks for a manual
  update of the running copy instead. When the location cannot be resolved, the update runs.
- **Time limits**: the periodic update's release lookup is cut off after 10 seconds (the
  required update keeps its lookup unbounded, as before, because it is the only way past the
  floor). The installer has no time
  limit, because killing it can leave Windows without `tloop.exe`; a stderr line announces the
  install so a slow download is visible.

### Version hold

`tloop update --to-version <version>` installs that release and records a version hold in
the per-user config directory (`TLOOP_CONFIG_DIR`, otherwise
`~/Library/Application Support/tloop` on macOS, `%APPDATA%\tloop` on Windows,
`$XDG_CONFIG_HOME/tloop` or `~/.config/tloop` elsewhere) as `dispatcher-version-hold.json`.
While the file exists — even if it cannot be read — the periodic self-update does not run.

- Plain `tloop update` removes the hold and updates to the newest release.
- Reinstalling with `install.sh` / `install.ps1` does not remove the hold.
- When a held dispatcher is below a project's `minimumDispatcherVersion`, the command fails
  with a manual-update error naming the hold instead of updating automatically. A Homebrew or
  winget install, or `TLOOP_DISABLE_SELF_UPDATE`, takes precedence and is the reason that error
  names instead.
