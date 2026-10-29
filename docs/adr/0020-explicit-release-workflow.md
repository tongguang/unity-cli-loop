# Explicit Releases Without Release Pull Requests

Date: 2026-10-29

## Decision

This personal Tuanjie fork uses one CI workflow and one manually dispatched
Release entry. The owner explicitly chooses a version increment. Go automation
writes version metadata and changelog entries, commits them to main, dispatches
checks for the exact commit, and invokes the three existing publishers.

This supersedes the release-please PR and GitHub App orchestration of ADR 0007
and the dispatcher post-publish stamp push of ADR 0005. The reason is an explicit
change in the fork's operating model: personal, intentional releases instead of
continuously maintained release PRs. It does not combine native and package
publication into a single commit. ADR 0007's hash-ordering invariant remains:
native artifacts must exist before a later package commit can pin their digests.

## Context

Every push used to refresh component PRs and dispatch duplicate checks. Bot PR
checks could wait for approval, while automatic merging and pin propagation
needed a separately installed App. This repository needs ordinary code checks
and owner-triggered releases; it does not use Homebrew, winget, Unity Cloud
signing, or hosted Editor testing.

## Consequences

The three existing publishing workflow paths remain signing identities trusted
by installed dispatchers. They are manually dispatched on main with an expected
SHA and correlation ID. No reusable-workflow conversion changes their identity.

The parent uses the ordinary workflow token, pushes without force, and explicitly
dispatches CI after bot commits. Every stage checks main still names its expected
commit. Native and package versions remain independent. Historical manifest,
config, contracts, pin fields, package identifiers and changelog history stay
compatible. Existing MIT notices remain.

The parent is serialized and stops on failure. Recovery rolls forward instead of
implementing a background retry/PR merge state machine. Actual Tuanjie Editor
validation remains local through the owner's check script, with affected tests.

## Reversal condition

Revisit when multiple independent maintainers need release PR review, branch
rules prohibit the version commits, or unattended release preparation becomes a
requirement. Do not add an App solely to reproduce the removed PR chain.
