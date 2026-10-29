# Package Manager Distribution

This Tuanjie fork distributes native CLI assets through GitHub Releases and the
UPM package through its package release tarball and Git URL. Installation and
upgrades use the published assets and attestation-verified pin.

Homebrew and winget publication are not configured. Their dormant workflow jobs
and token requirements were removed from the explicit release pipeline. Runtime
package-manager detection is unchanged. Reintroduce a distribution channel only
with a real repository, maintainer and credentials; release creation must remain
independent of an optional channel update.

See [Releasing](releasing.md) for the supported release entry.
