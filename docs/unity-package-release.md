# Tuanjie Package Release

This Tuanjie-only branch publishes an ordinary UPM tarball. Unity Cloud package
signing, a Unity organization, and an installed Editor are not required. This
policy was explicitly chosen for the Tuanjie fork; CLI archive attestations and
the verified dispatcher pin remain required.

`unity-package-release.yml` is the only workflow that publishes the package
release. The workflow builds the complete tarball before creating its tag and
release, so consumers never see a published release without its package asset.

## Flow

1. **Plan.** Completion of release-please, dispatcher-publish, or
   native-cli-publish starts the workflow on main. The plan reads the current
   release manifest, finds the package release commit, and checks that its runner
   and dispatcher releases and verified pin are ready. Missing dependencies are
   reported as a notice and checked again after their publish workflow completes.
2. **Pack.** A read-only job checks out the package release commit and copies
   `Packages/src` into the archive's `package/` directory. The tarball includes
   every source file, asset `.meta`, and skill in tilde-suffixed directories.
   `Packages/src/LICENSE.md` preserves the root project license and upstream
   copyright verbatim, alongside the existing third-party notices.
   No external package-signing binary or credentials run in this job.
3. **Stage.** `cmd/stage-unity-package` checks the package name, release version,
   and completeness against the source tree, then names the archive
   `io.github.tongguang.tloop-<version>.tgz`.
4. **Publish.** `cmd/create-unity-package-release` checks the release commit,
   tag, and changelog, then creates the release with its complete tarball. Only
   this job writes repository contents, and it runs only on `main`. Other branches
   may pack for validation. Published releases are never modified.
5. **Post-publish.** The workflow marks the package release PR tagged and starts
   release-please, because writes made with GITHUB_TOKEN do not start follow-up
   push workflows.

## Operations

- The manual `dry-run` option defaults to true. It packs the planned release
  source, or the tag of an already published release, and keeps the tarball as
  the `tloop-upm-package` Actions artifact for 14 days. It publishes nothing.
- To publish a ready release after repairing a failed run, dispatch the workflow
  on main with `dry-run=false`. It uses the manifest version and its release
  commit, rather than inventing a version or tagging a later commit.
- A failure before publication leaves no published package. Repair the cause
  and rerun the failed workflow. Do not publish a partial draft by hand.
- A published release missing its package asset must be replaced by a new
  version; the workflow will not patch a published release.
- This fork distributes through Git URLs and GitHub Release tarballs, without
  publishing to OpenUPM. `tloop package install` writes the Git dependency, using
  `main` by default or a release tag when `--version` is supplied.

No `upm-signing` environment or `UPM_*` secrets are used by this fork's package
workflow. The historical signed-package decision is retained in ADR 0013 with
the Tuanjie branch override recorded there.
