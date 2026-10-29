# Releasing Tuanjie CLI Loop

Normal pushes to `main` and pull requests run `CI`. They never prepare or publish
releases. Releases start only from the `Release` workflow on `main`.

## Before releasing

Commit and push the intended source and check its CI result. For C# changes and
upstream imports, run `scripts/check-tuanjie.ps1` against the open Tuanjie Editor
with the test filter that covers the change. Its default test class is a smoke
check, not a substitute for the affected tests. GitHub-hosted runners cannot run
this local Editor check; a green CI result covers the CLI and Editor-free tests.

## Start a release

In GitHub, open **Actions > Release > Run workflow**, select `main`, and choose
`patch`, `minor`, or `major`. The same entry is available through GitHub CLI:

```sh
gh workflow run release.yml --ref main -f bump=patch
```

Preview first without modifying source or publishing:

```sh
gh workflow run release.yml --ref main -f bump=patch -f dry-run=true
```

`components=auto` compares each component with its current version tag. It
includes the component root, shared CLI inputs, its packaging scripts and native
publishing workflow. Missing current tags cause a version advance, never a
rebuild of a historical tag. A dispatcher version not yet recorded in the package
pin also selects the package, including after an interrupted release. The
selection is deliberately conservative.
`components=all` forces all three components; `dispatcher` and `runner` also
include the package so its pin can ship the new native dependency. `package`
selects only the UPM package. The chosen increment applies independently to each
selected component's existing version series.

```sh
gh workflow run release.yml --ref main -f bump=minor -f components=all
gh run list --workflow release.yml
gh run watch <run-id> --exit-status
```

For a committed checkout, a local preview needs only Go and Git:

```sh
cd cli/release-automation
go run ./cmd/release --dry-run --repo tongguang/unity-cli-loop
```

The preview prints the source commit and proposed versions. It creates no files,
commits, tags, pull requests, releases, or child workflow runs. Live publication
is restricted to the GitHub workflow on `main`.

## What the workflow does

1. Determine changed components and generate their next versions and changelog
   entries. Existing changelog history is preserved. No release PR is opened.
2. When native components changed, commit their version metadata, push without
   force, explicitly dispatch CI for that exact commit, and wait for success.
3. Dispatch `dispatcher-publish.yml` and/or `native-cli-publish.yml` on that same
   commit. Those original workflows build, attest, verify and publish the native
   assets. They do not push pins or start other workflows.
4. Verify the current published dispatcher and stamp its attested asset manifest into
   the pin. Keep the package pin and project mirror byte-identical. Generate a
   later package-version commit containing that pin, then check it through CI.
5. Dispatch `unity-package-release.yml`, which verifies its dependencies, packs
   the complete source tree and creates the release with the verified tarball.
   Confirm the published tag points at the checked package commit.

The parent explicitly dispatches every child because `GITHUB_TOKEN` pushes do
not trigger ordinary push workflows. No GitHub App, extra PAT, release PR
approval, environment approval, or package-manager token is needed. The release
workflow needs Contents write and Actions write; CI keeps read-only source
access, with Security events write only for the CodeQL upload job.

Each child receives an expected SHA and a unique run ID including the parent
run attempt. A success from a previous run cannot satisfy the current release.
A moved `main`, failed check, wrong tag commit, invalid attestation, cancelled
run or missing package dependency stops the release. The workflow never force
pushes or moves an existing release tag.

## Failure and migration

An interrupted release can leave version commits or successfully published native
components on `main`. Start a new patch release to roll forward. Package-only
recovery also verifies and stamps a dispatcher published by an earlier run. Use
`components=all` when intentionally rolling all components forward. Never retag
an old native version to match a later build; see
[the recovery runbook](release-recovery-runbook.md).

The three historical tag formats and version series remain unchanged. The
historically named `.release-please-manifest.json` and `release-please-config.json`
remain metadata consumed by the package verifier; release-please no longer opens
PRs or stamps versions. Version fields are written only by `cmd/release` during
an explicit release. Do not edit them in feature changes.

After this change reaches `main`, close any still-open release-please PRs without
merging them. They belong to the retired pipeline. Repository branch rules that
require old workflow/job names must be updated to the unified CI jobs; this fork
currently has no such branch rules. Homebrew and winget publication are not
configured. MIT notices and the complete agent skills remain in every package.
