# Tuanjie UPM Package Release

`unity-package-release.yml` is the sole publisher of ordinary UPM tarballs for
this Tuanjie-only fork. It needs no Unity Cloud signing credentials. The normal
entry is [Release](releasing.md), which invokes it only after native dependencies
have been published and the later package-version commit has passed CI.

The publisher runs on main through `workflow_dispatch`. It validates its expected
source commit, finds the version commit from the manifest and changelog, checks
the protocol and pin, verifies published native dependencies, and packs the
complete `Packages/src` tree. `stage-unity-package` compares every archive entry
with that source tree before anything is published. CLI-only skills in tilde
folders and existing MIT notices remain in the archive.

The tarball is uploaded as a workflow artifact before publication. The release
is created with that tarball, so an immutable published release never waits for
its package asset. A direct manual run defaults to preview packing:

```sh
gh workflow run unity-package-release.yml --ref main -f dry-run=true
```

Use the unified Release entry for routine publication. Do not create UPM releases
by hand. On failure, follow [release recovery](release-recovery-runbook.md).
The package workflow starts no release-please run and updates no PR labels.
