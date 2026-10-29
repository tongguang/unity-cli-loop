# Dispatcher and Package Release Order

The current release entry is [Release](releasing.md). No component release PR is
created or merged. ADR 0020 supersedes the previous release-please/App mechanism;
the dependency order established by ADR 0007 still applies.

1. Commit and check native version metadata.
2. Publish selected native components from that exact main commit, keeping the
   signing workflows at their original paths.
3. Verify the dispatcher release attestation, then stamp `dispatcherReleaseTag`
   and `dispatcherArchiveManifest`. Mirror the complete pin byte-identically.
4. Generate a later UPM version commit that includes the verified pin, check it,
   then publish the package exclusively through `unity-package-release.yml`.

The package may never predict hashes for unpublished binaries or reuse a package
commit made before its new dispatcher assets existed. The verifier checks the
manifest version, package pin, protocol generation and published native assets.
`minimumDispatcherVersion` is raised only when the package needs a newer feature;
it is not automatically raised to the newest dispatcher version.

A package-only release keeps the published dispatcher pin. Native releases also
release a package so a fresh install uses the newly published native dependency.
If a stage fails, later stages stop; roll forward with a new explicit release.
