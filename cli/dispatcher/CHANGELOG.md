# Changelog

## [3.14.11](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.10...dispatcher-v3.14.11) (2026-11-11)

- fix: build native CLI with Go 1.27.4 (74d416f9)
- Merge upstream launch lock protection (46833f2e)
- Merge upstream compile recovery and attestation updates (b51aaad3)
- Merge upstream hot reload and launch updates while preserving Tuanjie behavior (9edd58ca)
- Merge upstream test memory fix while preserving Tuanjie behavior (23598d33)
- Merge upstream maintenance updates while preserving Tuanjie behavior (44e69026)
- fix: merge upstream main-thread blocked command handling (593411a8)
- Merge upstream Game View size and CLI stall notice fixes (3c80dde9)
- feat(verify-project): accept Tuanjie-encoded GUIDs (88b1a256)
- feat: retire the replay-input tool and disable input recording (2745e0b2)
- ci: simplify checks and add explicit releases (c404a0d7)
- fix(installer): reject raw carriage returns in pin manifests (ead482c5)
- fix: CLI binaries ship with the Go 1.27.3 runtime (d6060e4f)
- test: cover package manifest failure paths (e5e66112)
- fix: install Tuanjie package from Git and restrict releases to main (a2517f12)
- refactor: stop uninstalling an npm package this fork never published (c910fccf)
- refactor: drop the temporary V3 CLI invocation migration skill (20f97156)
- refactor(dispatcher): drop delegation to tloop V2 projects (db30fe6c)
- docs(skills): name Tuanjie where the remaining skill and help text said Unity (b95b8727)
- chore: finish the tloop rename in tools, scripts, comments, and docs (d34d29b1)
- fix: include upstream license in both native CLI modules (8f29e399)
- fix: document Tuanjie fork and preserve upstream licenses (6a661e9f)
- fix: prepare standalone tloop release automation (ca251796)
- feat: make tloop a Tuanjie-only CLI (77ee6912)
- fix: The uloop-status skill now covers cancelling a stuck command, replacing the separate uloop-cancel skill (#3404) (7cb3517e)
- fix: uloop no longer advertises verify-project in help, command lists, or installed skills (#3406) (76ce3f2e)
- feat: uloop launch --restart and --quit let Unity quit itself so it keeps the window layout (#3403) (57108dba)
- fix: uloop says the Unity Editor is not running instead of telling agents to wait and retry (#3400) (fb876dfd)
- chore(deps): bump golang.org/x/term from 0.46.0 to 0.47.0 in /cli/dispatcher (#3382) (927e70da)
- feat: Choose whether skills install at the Unity project root or the Git repository root (#3395) (a16951eb)
- feat: screenshot can render a prefab asset by path without opening it (#3383) (5556f240)
- chore(deps): bump golang.org/x/sys from 0.48.0 to 0.49.0 in /cli/dispatcher (#3337) (9664f273)
- chore(deps): bump github.com/Microsoft/go-winio from 0.6.2 to 0.6.4 in /cli/common (#3336) (31fb2d62)
- chore(deps): bump golang.org/x/sys from 0.48.0 to 0.49.0 in /cli/common (#3335) (538dfffd)
- revert: hide compile-check again until its place beside compile is settled (#3368) (87e69cd0)
- fix: CLI binaries ship with the patched Go 1.26.10 runtime (#3378) (0bfc974f)
- feat: record-video returns frame sheets and change numbers an agent can read, around one action in one command (#3377) (a8c665b4)
- fix: skill descriptions lead agents to pause-point and record-video when verifying a fix at runtime (#3371) (0e6d4832)
- feat: pick the simulated device by name with game-view-mode --device (#3370) (28f020b4)
- feat: game-view-mode shows or switches the Game view tab to the Device Simulator and matches its orientation (#3369) (24b46f1a)
- feat: compile-check closes its miss-side gaps and is shown as an ordinary command (#3364) (dfc73dae)
- feat: compile-check answers a method-body edit in a large assembly in seconds instead of a full recompile (#3318) (e917059f)
- feat: Screenshot annotation lists UI Toolkit controls, and simulate-mouse-ui clicks and drags them (#3310) (044c3337)
- feat: uloop cancel stops the tool the Editor is running, and the waiting CLI gets TOOL_CANCELLED (#3303) (cd65bb5b)
- fix: Hot reload no longer compiles just because a file without a baseline declares interface, abstract, or extern methods (#3299) (e8058b6f)
- ci: create the Unity package release only after its tarball is signed (#3270) (b04b58e8)
- fix: CLI binaries ship with the patched Go 1.26.9 runtime and golang.org/x/net 0.60.0 (#3261) (c676543c)
- fix: IL2CPP player builds no longer crash intermittently in an Editor started by uloop launch on macOS (#3210) (4532114f)
- feat: Hot reload waits for a busy Editor instead of compiling, stays fast as live patches grow, and pause points work in package scripts (#3224) (35084a8d)
- feat: Hot reload now applies edits on partial types and in global-using assemblies, and the response says whether an edit is live (#3198) (fed634ac)
- chore: Update the launch comment about how V3 schedules server recovery (#3172) (ffc9cf82)
- feat: uloop status reports whether Unity can take a command now, without side effects (#3163) (dbfd8b53)
- feat: uloop verify-project finds .meta, GUID, conflict-marker, and manifest problems without starting Unity (#3162) (bc15f81f)
- feat: Rerun only the tests that failed in the last run with run-tests --rerun-failed (#3161) (56718411)
- chore(deps): bump actions/download-artifact from 4.3.0 to 8.0.1 (#3059) (c6759f40)
- chore: Remove duplicated tests from the shared Go CLI module (#3152) (21237388)
- chore: Go CLI modules now have over 94% test coverage, more of the Unity package is under EditMode tests, and nightly runs track both (#3102) (470a8609)
- fix: Agents pick hot-reload and other long-described tools more easily with shorter skill descriptions (#3055) (b8cade79)
- fix: CLI Play no longer rewrites unsaved Scene files; --unsaved-changes chooses keep, save, or fail (#3048) (e85f53d3)
- chore: Homebrew and winget update failures no longer block the package pin and release (#3046) (35dd8cc7)
- fix: Timed-out Unity process lookups on Windows now report a timeout instead of a bare exit status (#3029) (1be51a3e)
- feat: Hot reload can introduce new types declared internal or without an access modifier, and they can use internal members of their assembly (#3017) (5333ae0d)
- feat: Hot reload can wire values into added fields without a compile, and pause points use the edited file's lines after a hot reload (#3008) (05bec468)
- fix: Syncing skills with --output-dir now skips tools the project disabled (#2997) (2e428f4c)
- chore: Windows CI tests no longer fail when PowerShell starts slowly (#2955) (a9f6b30c)
- fix: Busy errors no longer report a command as running while it only waits for a stalled Unity main thread (#2948) (0acc1a7d)
- fix: Busy errors explain that uloop runs one command at a time and how to recover (#2945) (e706b3f9)
- fix: Concurrent compile-check runs on one project wait for each other instead of corrupting each other's results (#2946) (6e2d1c5b)
- chore(deps): bump golang.org/x/term from 0.45.0 to 0.46.0 in /cli/dispatcher (#2809) (5f6b473b)
- chore: Go modules declare the 1.26.0 patch directive so dependency bumps that require it build (#2942) (98fdeead)
- chore: Keep the unreleased compile-check command out of help, command lists, and installed skills (#2919) (45c7f014)
- fix: The compile skill tells agents when to use hot-reload instead (#2917) (cbee05b9)
- feat: uloop installs, updates, and downloads its project runner on Linux (#2855) (828e1414)
- feat: uloop finds and launches the Unity Editor on Linux (#2854) (ba613489)
- feat: Linux (x86_64) builds of uloop are published with each release (#2852) (de3c5cfe)
- fix: A busy error while the Editor is paused says the running command cannot finish and how to recover (#2860) (63e4ac58)
- feat: Hot reload without --files selects again the new files a Play Mode domain reload discarded (#2853) (14c7cc5b)
- fix: Hot reload refuses edits that would drop earlier added methods, gives the right advice in Play Mode, and reports its results accurately (#2834) (6d4a57e3)
- fix: Hot reload reports what a run actually did and skips added methods it cannot verify (#2830) (c90d2265)
- feat: hot-reload says whether the CLI should compile the edits it could not apply, chosen by --compile-on-skip (#2819) (edf54a34)
- feat: uloop compile-check compiles C# with the bundled Roslyn without launching Unity (#2773) (dc0bb70b)
- fix(cli): launch --restart tolerates Temp files still held after Unity exits (#2766) (4e1fd86b)
- chore(deps): bump golang.org/x/sys from 0.47.0 to 0.48.0 in /cli/common (#2760) (4fdd1125)
- chore(deps): bump google.golang.org/grpc from 1.83.1 to 1.83.2 in /cli/dispatcher (#2729) (c17a003e)
- fix: installed skills no longer stay Outdated when a tool is turned off (#2757) (f24fe70b)
- fix: merge the Unity package release pull request only from automation (#2710) (1624b83b)
- chore: split release pull requests by component and auto-merge the package release (#2694) (c847ecb3)
- feat: enable-pause-point --persist re-arms the pause point after a domain reload (#2681) (9c0f1953)
- feat: record-video can now record any Editor window, not just the Game View (#2680) (902cb034)
- feat: Hot reload now applies newly written types and explains the declarations it cannot apply (#2676) (3bc94710)
- chore: pin the installed-skill freshness rule with a contract shared by the CLI and the Editor (#2665) (0bc77745)
- fix: skills list no longer reports installed skills as outdated when other tools leave files beside SKILL.md (#2664) (dcffdb78)
- chore: release main (#2651) (02d8358a)
- fix: find-game-objects and replay-input now document their option values in lowercase like the other single-word options (#2659) (07fd55cb)
- feat: uloop now uses your GitHub CLI login for GitHub API calls (#2656) (76348988)
- fix: uloop update now explains how to recover from a GitHub API rate limit (#2655) (a112f15e)
- feat: Record the Game View to video while Play Mode runs (#2487) (7cadda9d)
- chore: release main (#2610) (693e1147)
- fix: uloop --help and uloop list now read skill descriptions from packages referenced by a relative file: path (#2644) (57873627)
- fix: Hot reload keeps callers in other files bound to the newest added-member body (#2612) (5faf240c)
- feat: Hot reload now applies edits across several files of the same assembly, including members added in another edited file (#2608) (09662d9a)
- chore: Stable dispatcher releases now stamp the package pin on main without a pull request (#2602) (52b64f11)
- chore(deps): bump google.golang.org/grpc from 1.82.1 to 1.83.1 in /cli/dispatcher (#2597) (62213b2a)
- chore: release main (#2538) (a0620640)
- fix(hot-reload): report stale patches and stop silent patch loss on run-tests (#2588) (76b5af77)
- feat: Unity comes to the front faster on Windows (#2568) (499c72ef)
- fix(pause-point): accept --file/--line on clear-pause-point (#2585) (e32591cf)
- fix(pause-point): warn before clear-pause-point resumes Play Mode (#2570) (ba28ed9e)
- feat: Bring the Unity Editor to the front faster on macOS (#2569) (8dc9f74e)
- feat: enable-pause-point can capture values right after the line runs with --snapshot-timing post-line (#2543) (54aab031)
- feat: run-tests can filter by test class name (#2536) (de911932)
- fix: First-run runner download now explains GitHub API rate limits and how to recover (#2533) (fe3e1f3d)
- feat: run-tests can keep the project's Enter Play Mode settings and survive the Domain Reload with --respect-enter-play-mode-settings (#2529) (bec8e63d)
- chore: release main (#2504) (fe95c4b6)
- fix: embed VERSIONINFO into Windows CLI binaries and document Defender false positives (#2512) (01c219bc)
- fix(release-automation): skip CLI stamps for description-only catalog changes (#2502) (4c7b2c7f)
- chore: release main (#2499) (0b6d4544)
- fix: Skill instructions now match what the tools actually guarantee (#2498) (e839eefe)
- chore: release main (#2488) (f318be67)
- chore(deps): bump golang.org/x/sys from 0.46.0 to 0.47.0 in /cli/common (#2455) (9a312f27)
- chore(deps): bump github.com/sigstore/sigstore-go from 1.2.2 to 1.3.0 in /cli/dispatcher (#2458) (10a67a51)
- chore(deps): bump actions/checkout from 6.0.2 to 7.0.1 (#2456) (867254cb)
- feat(dispatcher): refuse self-update for winget-managed installs (e789c544)
- feat(ci): open winget-pkgs pull request after stable dispatcher publish (e2484981)
- chore: release main (#2477) (4ab28e6a)
- feat: run-tests can now keep, fail on, or discard unsaved editor changes (#2478) (758fb12b)
- chore: Stable dispatcher releases now open the pin-stamp pull request automatically (#2473) (517ad62e)
- chore: Installer error hints now point to SECURITY.md for the archive manifest override (#2471) (8484c5f2)
- chore: release main (#2468) (fd63dbc3)
- fix: show the dispatcher version for interactive uloop -v inside V2 and V3 projects (#2467) (9c794872)
- chore: release main (#2460) (36798c20)
- chore: merge v3-beta into main for the 3.0.0 stable release (b11e7beb)

## [3.14.10](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.9...dispatcher-v3.14.10) (2026-11-11)

- fix: build native CLI with Go 1.27.4 (74d416f9)
- Merge upstream launch lock protection (46833f2e)
- Merge upstream compile recovery and attestation updates (b51aaad3)
- Merge upstream hot reload and launch updates while preserving Tuanjie behavior (9edd58ca)
- Merge upstream test memory fix while preserving Tuanjie behavior (23598d33)
- Merge upstream maintenance updates while preserving Tuanjie behavior (44e69026)
- fix: merge upstream main-thread blocked command handling (593411a8)
- Merge upstream Game View size and CLI stall notice fixes (3c80dde9)
- feat(verify-project): accept Tuanjie-encoded GUIDs (88b1a256)
- feat: retire the replay-input tool and disable input recording (2745e0b2)
- ci: simplify checks and add explicit releases (c404a0d7)
- fix(installer): reject raw carriage returns in pin manifests (ead482c5)
- fix: CLI binaries ship with the Go 1.27.3 runtime (d6060e4f)
- test: cover package manifest failure paths (e5e66112)
- fix: install Tuanjie package from Git and restrict releases to main (a2517f12)
- refactor: stop uninstalling an npm package this fork never published (c910fccf)
- refactor: drop the temporary V3 CLI invocation migration skill (20f97156)
- refactor(dispatcher): drop delegation to tloop V2 projects (db30fe6c)
- docs(skills): name Tuanjie where the remaining skill and help text said Unity (b95b8727)
- chore: finish the tloop rename in tools, scripts, comments, and docs (d34d29b1)
- fix: include upstream license in both native CLI modules (8f29e399)
- fix: document Tuanjie fork and preserve upstream licenses (6a661e9f)
- fix: prepare standalone tloop release automation (ca251796)
- feat: make tloop a Tuanjie-only CLI (77ee6912)
- fix: The uloop-status skill now covers cancelling a stuck command, replacing the separate uloop-cancel skill (#3404) (7cb3517e)
- fix: uloop no longer advertises verify-project in help, command lists, or installed skills (#3406) (76ce3f2e)
- feat: uloop launch --restart and --quit let Unity quit itself so it keeps the window layout (#3403) (57108dba)
- fix: uloop says the Unity Editor is not running instead of telling agents to wait and retry (#3400) (fb876dfd)
- chore(deps): bump golang.org/x/term from 0.46.0 to 0.47.0 in /cli/dispatcher (#3382) (927e70da)
- feat: Choose whether skills install at the Unity project root or the Git repository root (#3395) (a16951eb)
- feat: screenshot can render a prefab asset by path without opening it (#3383) (5556f240)
- chore(deps): bump golang.org/x/sys from 0.48.0 to 0.49.0 in /cli/dispatcher (#3337) (9664f273)
- chore(deps): bump github.com/Microsoft/go-winio from 0.6.2 to 0.6.4 in /cli/common (#3336) (31fb2d62)
- chore(deps): bump golang.org/x/sys from 0.48.0 to 0.49.0 in /cli/common (#3335) (538dfffd)
- revert: hide compile-check again until its place beside compile is settled (#3368) (87e69cd0)
- fix: CLI binaries ship with the patched Go 1.26.10 runtime (#3378) (0bfc974f)
- feat: record-video returns frame sheets and change numbers an agent can read, around one action in one command (#3377) (a8c665b4)
- fix: skill descriptions lead agents to pause-point and record-video when verifying a fix at runtime (#3371) (0e6d4832)
- feat: pick the simulated device by name with game-view-mode --device (#3370) (28f020b4)
- feat: game-view-mode shows or switches the Game view tab to the Device Simulator and matches its orientation (#3369) (24b46f1a)
- feat: compile-check closes its miss-side gaps and is shown as an ordinary command (#3364) (dfc73dae)
- feat: compile-check answers a method-body edit in a large assembly in seconds instead of a full recompile (#3318) (e917059f)
- feat: Screenshot annotation lists UI Toolkit controls, and simulate-mouse-ui clicks and drags them (#3310) (044c3337)
- feat: uloop cancel stops the tool the Editor is running, and the waiting CLI gets TOOL_CANCELLED (#3303) (cd65bb5b)
- fix: Hot reload no longer compiles just because a file without a baseline declares interface, abstract, or extern methods (#3299) (e8058b6f)
- ci: create the Unity package release only after its tarball is signed (#3270) (b04b58e8)
- fix: CLI binaries ship with the patched Go 1.26.9 runtime and golang.org/x/net 0.60.0 (#3261) (c676543c)
- fix: IL2CPP player builds no longer crash intermittently in an Editor started by uloop launch on macOS (#3210) (4532114f)
- feat: Hot reload waits for a busy Editor instead of compiling, stays fast as live patches grow, and pause points work in package scripts (#3224) (35084a8d)
- feat: Hot reload now applies edits on partial types and in global-using assemblies, and the response says whether an edit is live (#3198) (fed634ac)
- chore: Update the launch comment about how V3 schedules server recovery (#3172) (ffc9cf82)
- feat: uloop status reports whether Unity can take a command now, without side effects (#3163) (dbfd8b53)
- feat: uloop verify-project finds .meta, GUID, conflict-marker, and manifest problems without starting Unity (#3162) (bc15f81f)
- feat: Rerun only the tests that failed in the last run with run-tests --rerun-failed (#3161) (56718411)
- chore(deps): bump actions/download-artifact from 4.3.0 to 8.0.1 (#3059) (c6759f40)
- chore: Remove duplicated tests from the shared Go CLI module (#3152) (21237388)
- chore: Go CLI modules now have over 94% test coverage, more of the Unity package is under EditMode tests, and nightly runs track both (#3102) (470a8609)
- fix: Agents pick hot-reload and other long-described tools more easily with shorter skill descriptions (#3055) (b8cade79)
- fix: CLI Play no longer rewrites unsaved Scene files; --unsaved-changes chooses keep, save, or fail (#3048) (e85f53d3)
- chore: Homebrew and winget update failures no longer block the package pin and release (#3046) (35dd8cc7)
- fix: Timed-out Unity process lookups on Windows now report a timeout instead of a bare exit status (#3029) (1be51a3e)
- feat: Hot reload can introduce new types declared internal or without an access modifier, and they can use internal members of their assembly (#3017) (5333ae0d)
- feat: Hot reload can wire values into added fields without a compile, and pause points use the edited file's lines after a hot reload (#3008) (05bec468)
- fix: Syncing skills with --output-dir now skips tools the project disabled (#2997) (2e428f4c)
- chore: Windows CI tests no longer fail when PowerShell starts slowly (#2955) (a9f6b30c)
- fix: Busy errors no longer report a command as running while it only waits for a stalled Unity main thread (#2948) (0acc1a7d)
- fix: Busy errors explain that uloop runs one command at a time and how to recover (#2945) (e706b3f9)
- fix: Concurrent compile-check runs on one project wait for each other instead of corrupting each other's results (#2946) (6e2d1c5b)
- chore(deps): bump golang.org/x/term from 0.45.0 to 0.46.0 in /cli/dispatcher (#2809) (5f6b473b)
- chore: Go modules declare the 1.26.0 patch directive so dependency bumps that require it build (#2942) (98fdeead)
- chore: Keep the unreleased compile-check command out of help, command lists, and installed skills (#2919) (45c7f014)
- fix: The compile skill tells agents when to use hot-reload instead (#2917) (cbee05b9)
- feat: uloop installs, updates, and downloads its project runner on Linux (#2855) (828e1414)
- feat: uloop finds and launches the Unity Editor on Linux (#2854) (ba613489)
- feat: Linux (x86_64) builds of uloop are published with each release (#2852) (de3c5cfe)
- fix: A busy error while the Editor is paused says the running command cannot finish and how to recover (#2860) (63e4ac58)
- feat: Hot reload without --files selects again the new files a Play Mode domain reload discarded (#2853) (14c7cc5b)
- fix: Hot reload refuses edits that would drop earlier added methods, gives the right advice in Play Mode, and reports its results accurately (#2834) (6d4a57e3)
- fix: Hot reload reports what a run actually did and skips added methods it cannot verify (#2830) (c90d2265)
- feat: hot-reload says whether the CLI should compile the edits it could not apply, chosen by --compile-on-skip (#2819) (edf54a34)
- feat: uloop compile-check compiles C# with the bundled Roslyn without launching Unity (#2773) (dc0bb70b)
- fix(cli): launch --restart tolerates Temp files still held after Unity exits (#2766) (4e1fd86b)
- chore(deps): bump golang.org/x/sys from 0.47.0 to 0.48.0 in /cli/common (#2760) (4fdd1125)
- chore(deps): bump google.golang.org/grpc from 1.83.1 to 1.83.2 in /cli/dispatcher (#2729) (c17a003e)
- fix: installed skills no longer stay Outdated when a tool is turned off (#2757) (f24fe70b)
- fix: merge the Unity package release pull request only from automation (#2710) (1624b83b)
- chore: split release pull requests by component and auto-merge the package release (#2694) (c847ecb3)
- feat: enable-pause-point --persist re-arms the pause point after a domain reload (#2681) (9c0f1953)
- feat: record-video can now record any Editor window, not just the Game View (#2680) (902cb034)
- feat: Hot reload now applies newly written types and explains the declarations it cannot apply (#2676) (3bc94710)
- chore: pin the installed-skill freshness rule with a contract shared by the CLI and the Editor (#2665) (0bc77745)
- fix: skills list no longer reports installed skills as outdated when other tools leave files beside SKILL.md (#2664) (dcffdb78)
- chore: release main (#2651) (02d8358a)
- fix: find-game-objects and replay-input now document their option values in lowercase like the other single-word options (#2659) (07fd55cb)
- feat: uloop now uses your GitHub CLI login for GitHub API calls (#2656) (76348988)
- fix: uloop update now explains how to recover from a GitHub API rate limit (#2655) (a112f15e)
- feat: Record the Game View to video while Play Mode runs (#2487) (7cadda9d)
- chore: release main (#2610) (693e1147)
- fix: uloop --help and uloop list now read skill descriptions from packages referenced by a relative file: path (#2644) (57873627)
- fix: Hot reload keeps callers in other files bound to the newest added-member body (#2612) (5faf240c)
- feat: Hot reload now applies edits across several files of the same assembly, including members added in another edited file (#2608) (09662d9a)
- chore: Stable dispatcher releases now stamp the package pin on main without a pull request (#2602) (52b64f11)
- chore(deps): bump google.golang.org/grpc from 1.82.1 to 1.83.1 in /cli/dispatcher (#2597) (62213b2a)
- chore: release main (#2538) (a0620640)
- fix(hot-reload): report stale patches and stop silent patch loss on run-tests (#2588) (76b5af77)
- feat: Unity comes to the front faster on Windows (#2568) (499c72ef)
- fix(pause-point): accept --file/--line on clear-pause-point (#2585) (e32591cf)
- fix(pause-point): warn before clear-pause-point resumes Play Mode (#2570) (ba28ed9e)
- feat: Bring the Unity Editor to the front faster on macOS (#2569) (8dc9f74e)
- feat: enable-pause-point can capture values right after the line runs with --snapshot-timing post-line (#2543) (54aab031)
- feat: run-tests can filter by test class name (#2536) (de911932)
- fix: First-run runner download now explains GitHub API rate limits and how to recover (#2533) (fe3e1f3d)
- feat: run-tests can keep the project's Enter Play Mode settings and survive the Domain Reload with --respect-enter-play-mode-settings (#2529) (bec8e63d)
- chore: release main (#2504) (fe95c4b6)
- fix: embed VERSIONINFO into Windows CLI binaries and document Defender false positives (#2512) (01c219bc)
- fix(release-automation): skip CLI stamps for description-only catalog changes (#2502) (4c7b2c7f)
- chore: release main (#2499) (0b6d4544)
- fix: Skill instructions now match what the tools actually guarantee (#2498) (e839eefe)
- chore: release main (#2488) (f318be67)
- chore(deps): bump golang.org/x/sys from 0.46.0 to 0.47.0 in /cli/common (#2455) (9a312f27)
- chore(deps): bump github.com/sigstore/sigstore-go from 1.2.2 to 1.3.0 in /cli/dispatcher (#2458) (10a67a51)
- chore(deps): bump actions/checkout from 6.0.2 to 7.0.1 (#2456) (867254cb)
- feat(dispatcher): refuse self-update for winget-managed installs (e789c544)
- feat(ci): open winget-pkgs pull request after stable dispatcher publish (e2484981)
- chore: release main (#2477) (4ab28e6a)
- feat: run-tests can now keep, fail on, or discard unsaved editor changes (#2478) (758fb12b)
- chore: Stable dispatcher releases now open the pin-stamp pull request automatically (#2473) (517ad62e)
- chore: Installer error hints now point to SECURITY.md for the archive manifest override (#2471) (8484c5f2)
- chore: release main (#2468) (fd63dbc3)
- fix: show the dispatcher version for interactive uloop -v inside V2 and V3 projects (#2467) (9c794872)
- chore: release main (#2460) (36798c20)
- chore: merge v3-beta into main for the 3.0.0 stable release (b11e7beb)

## [3.14.9](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.8...dispatcher-v3.14.9) (2026-11-11)

- fix: build native CLI with Go 1.27.4 (74d416f9)
- Merge upstream launch lock protection (46833f2e)

## [3.14.8](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.7...dispatcher-v3.14.8) (2026-11-09)

- fix: uloop compile recovers when Unity accepts a compile but never starts it in the background (#3453) (cafc5640)
- fix: CLI binaries ship with Go 1.27.3 and sigstore-go 1.4.0, and now require macOS 13 or later (#3447) (6ee816a5)

## [3.14.7](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.6...dispatcher-v3.14.7) (2026-11-06)

- fix: A project path containing "-batchmode" no longer makes uloop lose the running Editor and open a second one (#3442) (27177f23)
- feat: uloop launch can give each Editor its own log file (#3438) (5273f3f2)

## [3.14.6](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.5...dispatcher-v3.14.6) (2026-11-06)

- fix: Repeated test runs no longer grow Unity's memory through log stack traces (#3429) (527ade67)
- chore: Keep the product name in one constant for the Editor and the uloop CLI (#3425) (0c9ecbd4)
- chore(deps): bump actions/create-github-app-token from 3.2.0 to 3.3.0 (#3423) (3fc98ba9)
- chore(deps): bump actions/upload-artifact from 7.0.1 to 7.0.2 (#3422) (864738d8)
- chore(deps): bump actions/download-artifact from 8.0.1 to 8.0.2 (#3421) (65b60fff)

## [3.14.5](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.4...dispatcher-v3.14.5) (2026-10-29)

- fix: Commands sent while a dialog blocks the Unity Editor now fail fast with MAIN_THREAD_BLOCKED instead of waiting silently (#3417) (1dfaeae5)

## [3.14.4](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.3...dispatcher-v3.14.4) (2026-10-29)

- fix: Agents now see the Unity main-thread stall notice even when stderr is not a terminal (#3416) (eba0ea48)
- fix: set-game-view-size no longer reports success under the Device Simulator, and --reset removes the size entry it leaves behind (#3412) (4ecde684)

## [3.14.3](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.2...dispatcher-v3.14.3) (2026-10-29)

- feat(verify-project): accept Tuanjie-encoded GUIDs (88b1a256)
- feat: retire the replay-input tool and disable input recording (2745e0b2)
- ci: simplify checks and add explicit releases (c404a0d7)
- fix(installer): reject raw carriage returns in pin manifests (ead482c5)
- fix: CLI binaries ship with the Go 1.27.3 runtime (d6060e4f)

## [3.14.2](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.1...dispatcher-v3.14.2) (2026-10-29)


### Bug Fixes

* install Tuanjie package from Git and restrict releases to main ([a2517f1](https://github.com/tongguang/unity-cli-loop/commit/a2517f12183d7d0aeca133474d0b076ded3ce51c))

## [3.14.1](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.14.0...dispatcher-v3.14.1) (2026-10-28)


### Bug Fixes

* include upstream license in both native CLI modules ([8f29e39](https://github.com/tongguang/unity-cli-loop/commit/8f29e399adcf55e82f91f2ac7354e0ead2b50b2d))

## [3.14.0](https://github.com/tongguang/unity-cli-loop/compare/dispatcher-v3.13.1...dispatcher-v3.14.0) (2026-10-28)


### Features

* make tloop a Tuanjie-only CLI ([77ee691](https://github.com/tongguang/unity-cli-loop/commit/77ee6912d7641709737d277549b0fc8efc1e6daf))

## [3.13.1](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.13.0...dispatcher-v3.13.1) (2026-10-27)


### Bug Fixes

* The uloop-status skill now covers cancelling a stuck command, replacing the separate uloop-cancel skill ([#3404](https://github.com/hatayama/unity-cli-loop/issues/3404)) ([7cb3517](https://github.com/hatayama/unity-cli-loop/commit/7cb3517ed7726d01bf017d521566153205ca6ec5))

## [3.13.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.12.0...dispatcher-v3.13.0) (2026-10-27)


### Features

* uloop launch --restart and --quit let Unity quit itself so it keeps the window layout ([#3403](https://github.com/hatayama/unity-cli-loop/issues/3403)) ([57108db](https://github.com/hatayama/unity-cli-loop/commit/57108dbaee73cc00703132451a65e792795bec62))


### Bug Fixes

* uloop no longer advertises verify-project in help, command lists, or installed skills ([#3406](https://github.com/hatayama/unity-cli-loop/issues/3406)) ([76ce3f2](https://github.com/hatayama/unity-cli-loop/commit/76ce3f2e30b1064a708be8d1709f4b452de3c31c))
* uloop says the Unity Editor is not running instead of telling agents to wait and retry ([#3400](https://github.com/hatayama/unity-cli-loop/issues/3400)) ([fb876df](https://github.com/hatayama/unity-cli-loop/commit/fb876dfd50fef92b82f4a731a4ee4216d20e79c0))

## [3.12.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.11.0...dispatcher-v3.12.0) (2026-10-27)


### Features

* Choose whether skills install at the Unity project root or the Git repository root ([#3395](https://github.com/hatayama/unity-cli-loop/issues/3395)) ([a16951e](https://github.com/hatayama/unity-cli-loop/commit/a16951ebb371e5231d5d22d15d06320efec3ad7b))

## [3.11.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.10.0...dispatcher-v3.11.0) (2026-10-22)


### Features

* screenshot can render a prefab asset by path without opening it ([#3383](https://github.com/hatayama/unity-cli-loop/issues/3383)) ([5556f24](https://github.com/hatayama/unity-cli-loop/commit/5556f24031ba4f63028dbfa4ce233b503ff1d2a2))

## [3.10.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.9.0...dispatcher-v3.10.0) (2026-10-21)


### Features

* compile-check answers a method-body edit in a large assembly in seconds instead of a full recompile ([#3318](https://github.com/hatayama/unity-cli-loop/issues/3318)) ([e917059](https://github.com/hatayama/unity-cli-loop/commit/e917059fde46431a24a9f4c7d84ca3e686160307))
* compile-check closes its miss-side gaps and is shown as an ordinary command ([#3364](https://github.com/hatayama/unity-cli-loop/issues/3364)) ([dfc73da](https://github.com/hatayama/unity-cli-loop/commit/dfc73dae8c898318baa34b32a370ffca7c7344cc))
* game-view-mode shows or switches the Game view tab to the Device Simulator and matches its orientation ([#3369](https://github.com/hatayama/unity-cli-loop/issues/3369)) ([24b46f1](https://github.com/hatayama/unity-cli-loop/commit/24b46f1ac4eb6be7b7776608d438461a1a7fa5bd))
* pick the simulated device by name with game-view-mode --device ([#3370](https://github.com/hatayama/unity-cli-loop/issues/3370)) ([28f020b](https://github.com/hatayama/unity-cli-loop/commit/28f020b4598b6ab2ae9599cec9eca53f9424fcad))
* record-video returns frame sheets and change numbers an agent can read, around one action in one command ([#3377](https://github.com/hatayama/unity-cli-loop/issues/3377)) ([a8c665b](https://github.com/hatayama/unity-cli-loop/commit/a8c665b47006922a14e743b0846b26b2ae1c3138))


### Reverts

* hide compile-check again until its place beside compile is settled ([#3368](https://github.com/hatayama/unity-cli-loop/issues/3368)) ([87e69cd](https://github.com/hatayama/unity-cli-loop/commit/87e69cd0508d9ac14cb13c2544376b60fa037b64))

## [3.9.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.8.2...dispatcher-v3.9.0) (2026-10-13)


### Features

* uloop cancel stops the tool the Editor is running, and the waiting CLI gets TOOL_CANCELLED ([#3303](https://github.com/hatayama/unity-cli-loop/issues/3303)) ([cd65bb5](https://github.com/hatayama/unity-cli-loop/commit/cd65bb5b78aed5d57bc4687ce573cd007266ee59))

## [3.8.2](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.8.1...dispatcher-v3.8.2) (2026-10-10)


### Bug Fixes

* CLI binaries ship with the patched Go 1.26.9 runtime and golang.org/x/net 0.60.0 ([#3261](https://github.com/hatayama/unity-cli-loop/issues/3261)) ([c676543](https://github.com/hatayama/unity-cli-loop/commit/c676543c11fdb61f21bf1c0b2bfb6231473651b1))

## [3.8.1](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.8.0...dispatcher-v3.8.1) (2026-10-07)


### Bug Fixes

* IL2CPP player builds no longer crash intermittently in an Editor started by uloop launch on macOS ([#3210](https://github.com/hatayama/unity-cli-loop/issues/3210)) ([4532114](https://github.com/hatayama/unity-cli-loop/commit/4532114f87f224a649a71a1d1cd68e20563ccb42))

## [3.8.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.7.0...dispatcher-v3.8.0) (2026-10-07)


### Features

* Hot reload now applies edits on partial types and in global-using assemblies, and the response says whether an edit is live ([#3198](https://github.com/hatayama/unity-cli-loop/issues/3198)) ([fed634a](https://github.com/hatayama/unity-cli-loop/commit/fed634ac769f999ee53f6453491e9dd7e95214c5))

## [3.7.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.6.3...dispatcher-v3.7.0) (2026-10-06)


### Features

* Rerun only the tests that failed in the last run with run-tests --rerun-failed ([#3161](https://github.com/hatayama/unity-cli-loop/issues/3161)) ([5671841](https://github.com/hatayama/unity-cli-loop/commit/56718411a3a3d8a8f217b2a7b274745f9c0a6d3f))
* uloop status reports whether Unity can take a command now, without side effects ([#3163](https://github.com/hatayama/unity-cli-loop/issues/3163)) ([dbfd8b5](https://github.com/hatayama/unity-cli-loop/commit/dbfd8b538d7d0ac23b703b5793fb3fd3184ea2d9))
* uloop verify-project finds .meta, GUID, conflict-marker, and manifest problems without starting Unity ([#3162](https://github.com/hatayama/unity-cli-loop/issues/3162)) ([bc15f81](https://github.com/hatayama/unity-cli-loop/commit/bc15f81f07a360a0248d84bc34d1445b76263ea4))

## [3.6.3](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.6.2...dispatcher-v3.6.3) (2026-09-30)


### Bug Fixes

* CLI Play no longer rewrites unsaved Scene files; --unsaved-changes chooses keep, save, or fail ([#3048](https://github.com/hatayama/unity-cli-loop/issues/3048)) ([e85f53d](https://github.com/hatayama/unity-cli-loop/commit/e85f53d30809f3cbe55a9e918a568bc9fd322bcd))

## [3.6.2](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.6.1...dispatcher-v3.6.2) (2026-09-30)


### Bug Fixes

* Timed-out Unity process lookups on Windows now report a timeout instead of a bare exit status ([#3029](https://github.com/hatayama/unity-cli-loop/issues/3029)) ([1be51a3](https://github.com/hatayama/unity-cli-loop/commit/1be51a3eb1f6bda7a1cfc524728f02f30a85bb8c))

## [3.6.1](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.6.0...dispatcher-v3.6.1) (2026-09-25)


### Bug Fixes

* Busy errors explain that uloop runs one command at a time and how to recover ([#2945](https://github.com/hatayama/unity-cli-loop/issues/2945)) ([e706b3f](https://github.com/hatayama/unity-cli-loop/commit/e706b3f92f6c054ed9b1b440837bc9fad2262f15))
* Busy errors no longer report a command as running while it only waits for a stalled Unity main thread ([#2948](https://github.com/hatayama/unity-cli-loop/issues/2948)) ([0acc1a7](https://github.com/hatayama/unity-cli-loop/commit/0acc1a7d652dd64ad5fc9346346852f5160758a5))
* Concurrent compile-check runs on one project wait for each other instead of corrupting each other's results ([#2946](https://github.com/hatayama/unity-cli-loop/issues/2946)) ([6e2d1c5](https://github.com/hatayama/unity-cli-loop/commit/6e2d1c5ba3df51929963a0c5c6e963e716ba3213))
* Syncing skills with --output-dir now skips tools the project disabled ([#2997](https://github.com/hatayama/unity-cli-loop/issues/2997)) ([2e428f4](https://github.com/hatayama/unity-cli-loop/commit/2e428f4c2f47c07d321c50eba9d20652b6bcf280))

## [3.6.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.5.1...dispatcher-v3.6.0) (2026-09-23)


### Features

* hot-reload says whether the CLI should compile the edits it could not apply, chosen by --compile-on-skip ([#2819](https://github.com/hatayama/unity-cli-loop/issues/2819)) ([edf54a3](https://github.com/hatayama/unity-cli-loop/commit/edf54a34a1440c5a130477f406248033ea977d88))
* uloop finds and launches the Unity Editor on Linux ([#2854](https://github.com/hatayama/unity-cli-loop/issues/2854)) ([ba61348](https://github.com/hatayama/unity-cli-loop/commit/ba613489bbe2f65b0807859562e8720a75f4300c))
* uloop installs, updates, and downloads its project runner on Linux ([#2855](https://github.com/hatayama/unity-cli-loop/issues/2855)) ([828e141](https://github.com/hatayama/unity-cli-loop/commit/828e14144ef0b100a7c7e099370bd824c2a2d103))


### Bug Fixes

* A busy error while the Editor is paused says the running command cannot finish and how to recover ([#2860](https://github.com/hatayama/unity-cli-loop/issues/2860)) ([63e4ac5](https://github.com/hatayama/unity-cli-loop/commit/63e4ac58b8b18ce1f83ea00ba8c57e9a8159ec51))
* **cli:** launch --restart tolerates Temp files still held after Unity exits ([#2766](https://github.com/hatayama/unity-cli-loop/issues/2766)) ([4e1fd86](https://github.com/hatayama/unity-cli-loop/commit/4e1fd86bbfa96986df104384c2e0e53a593a206b))
* Hot reload reports what a run actually did and skips added methods it cannot verify ([#2830](https://github.com/hatayama/unity-cli-loop/issues/2830)) ([c90d226](https://github.com/hatayama/unity-cli-loop/commit/c90d2265a7f5f1bdafe958c24e4bf26aa44f69f4))

## [3.5.1](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.5.0...dispatcher-v3.5.1) (2026-09-10)


### Bug Fixes

* installed skills no longer stay Outdated when a tool is turned off ([#2757](https://github.com/hatayama/unity-cli-loop/issues/2757)) ([f24fe70](https://github.com/hatayama/unity-cli-loop/commit/f24fe70b180c7647e2ac80e92af777b98c4ce362))

## [3.5.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.4.0...dispatcher-v3.5.0) (2026-09-08)


### Features

* enable-pause-point --persist re-arms the pause point after a domain reload ([#2681](https://github.com/hatayama/unity-cli-loop/issues/2681)) ([9c0f195](https://github.com/hatayama/unity-cli-loop/commit/9c0f195354785295e2d1e1d8887d0ad37c6c4414))
* record-video can now record any Editor window, not just the Game View ([#2680](https://github.com/hatayama/unity-cli-loop/issues/2680)) ([902cb03](https://github.com/hatayama/unity-cli-loop/commit/902cb034ca57ceea7c18fc794cc09bd005471a05))


### Bug Fixes

* skills list no longer reports installed skills as outdated when other tools leave files beside SKILL.md ([#2664](https://github.com/hatayama/unity-cli-loop/issues/2664)) ([dcffdb7](https://github.com/hatayama/unity-cli-loop/commit/dcffdb787d18532d12385a17985ff29fed955a37))

## [3.4.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.3.1...dispatcher-v3.4.0) (2026-09-07)


### Features

* Record the Game View to video while Play Mode runs ([#2487](https://github.com/hatayama/unity-cli-loop/issues/2487)) ([7cadda9](https://github.com/hatayama/unity-cli-loop/commit/7cadda9d1fb451d0fcafe33d8f32929049937d16))
* uloop now uses your GitHub CLI login for GitHub API calls ([#2656](https://github.com/hatayama/unity-cli-loop/issues/2656)) ([7634898](https://github.com/hatayama/unity-cli-loop/commit/76348988f4b82a61b0225384e890c5023f5ec9a0))


### Bug Fixes

* find-game-objects and replay-input now document their option values in lowercase like the other single-word options ([#2659](https://github.com/hatayama/unity-cli-loop/issues/2659)) ([07fd55c](https://github.com/hatayama/unity-cli-loop/commit/07fd55cbd484ee7235160c246078ca166622c6f3))
* uloop update now explains how to recover from a GitHub API rate limit ([#2655](https://github.com/hatayama/unity-cli-loop/issues/2655)) ([a112f15](https://github.com/hatayama/unity-cli-loop/commit/a112f15e85b9150e58b34147300fb16b4fdeb19a))

## [3.3.1](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.3.0...dispatcher-v3.3.1) (2026-09-06)


### Bug Fixes

* uloop --help and uloop list now read skill descriptions from packages referenced by a relative file: path ([#2644](https://github.com/hatayama/unity-cli-loop/issues/2644)) ([5787362](https://github.com/hatayama/unity-cli-loop/commit/57873627bda7250f711a23757255c52109843f72))

## [3.3.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.2.2...dispatcher-v3.3.0) (2026-09-04)


### Features

* Bring the Unity Editor to the front faster on macOS ([#2569](https://github.com/hatayama/unity-cli-loop/issues/2569)) ([8dc9f74](https://github.com/hatayama/unity-cli-loop/commit/8dc9f74e9203f2e61ae3eeb2e5f7980851a7c16e))
* enable-pause-point can capture values right after the line runs with --snapshot-timing post-line ([#2543](https://github.com/hatayama/unity-cli-loop/issues/2543)) ([54aab03](https://github.com/hatayama/unity-cli-loop/commit/54aab031ccbc1fb980f77a10edad499f018c05d2))
* run-tests can filter by test class name ([#2536](https://github.com/hatayama/unity-cli-loop/issues/2536)) ([de91193](https://github.com/hatayama/unity-cli-loop/commit/de911932f26894f34bc9a26f1b6cc79e5c4a1a89))
* run-tests can keep the project's Enter Play Mode settings and survive the Domain Reload with --respect-enter-play-mode-settings ([#2529](https://github.com/hatayama/unity-cli-loop/issues/2529)) ([bec8e63](https://github.com/hatayama/unity-cli-loop/commit/bec8e63d7c41288178c8be56af8c2156f9eb1af7))
* Unity comes to the front faster on Windows ([#2568](https://github.com/hatayama/unity-cli-loop/issues/2568)) ([499c72e](https://github.com/hatayama/unity-cli-loop/commit/499c72ef21ff09ac1e8eba2b4d838e2b5fee0074))


### Bug Fixes

* First-run runner download now explains GitHub API rate limits and how to recover ([#2533](https://github.com/hatayama/unity-cli-loop/issues/2533)) ([fe3e1f3](https://github.com/hatayama/unity-cli-loop/commit/fe3e1f3d2fc180722480059018cc26068c567a8f))
* **hot-reload:** report stale patches and stop silent patch loss on run-tests ([#2588](https://github.com/hatayama/unity-cli-loop/issues/2588)) ([76b5af7](https://github.com/hatayama/unity-cli-loop/commit/76b5af777ac1891f906806907d44aaceecfb02a8))
* **pause-point:** accept --file/--line on clear-pause-point ([#2585](https://github.com/hatayama/unity-cli-loop/issues/2585)) ([e32591c](https://github.com/hatayama/unity-cli-loop/commit/e32591cfad069704c8f1d63dd74baf16fddbe0eb))

## [3.2.2](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.2.1...dispatcher-v3.2.2) (2026-09-02)


### Bug Fixes

* embed VERSIONINFO into Windows CLI binaries and document Defender false positives ([#2512](https://github.com/hatayama/unity-cli-loop/issues/2512)) ([01c219b](https://github.com/hatayama/unity-cli-loop/commit/01c219bcb16380ec4ff0d8216be394221cf0221b))
* **release-automation:** skip CLI stamps for description-only catalog changes ([#2502](https://github.com/hatayama/unity-cli-loop/issues/2502)) ([4c7b2c7](https://github.com/hatayama/unity-cli-loop/commit/4c7b2c7f11e86cb882135bb54429bd56bc5c3233))

## [3.2.1](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.2.0...dispatcher-v3.2.1) (2026-09-02)


### Bug Fixes

* Skill instructions now match what the tools actually guarantee ([#2498](https://github.com/hatayama/unity-cli-loop/issues/2498)) ([e839eef](https://github.com/hatayama/unity-cli-loop/commit/e839eefe29900eb370d8522fb2979c9e0db36127))

## [3.2.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.1.0...dispatcher-v3.2.0) (2026-09-02)


### Features

* **dispatcher:** refuse self-update for winget-managed installs ([e789c54](https://github.com/hatayama/unity-cli-loop/commit/e789c544fea774136b85cdaa6b496bbb6c4e03ab))

## [3.1.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.1...dispatcher-v3.1.0) (2026-08-31)


### Features

* run-tests can now keep, fail on, or discard unsaved editor changes ([#2478](https://github.com/hatayama/unity-cli-loop/issues/2478)) ([758fb12](https://github.com/hatayama/unity-cli-loop/commit/758fb12bc0c9ea7f199f59bcbfa6002485d6b41b))

## [3.0.1](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0...dispatcher-v3.0.1) (2026-08-31)


### Bug Fixes

* show the dispatcher version for interactive uloop -v inside V2 and V3 projects ([#2467](https://github.com/hatayama/unity-cli-loop/issues/2467)) ([9c79487](https://github.com/hatayama/unity-cli-loop/commit/9c79487224624bf52011867c5f57fbb9130fb283))

## [3.0.0](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.37...dispatcher-v3.0.0) (2026-08-31)


### Miscellaneous Chores

* force release 3.0.0 ([6233aa5](https://github.com/hatayama/unity-cli-loop/commit/6233aa514a2fb63103ae5780a42602de30becd28))

## [3.0.0-beta.37](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.36...dispatcher-v3.0.0-beta.37) (2026-08-28)


### Bug Fixes

* package install now finds the Unity project from the repository root ([#2448](https://github.com/hatayama/unity-cli-loop/issues/2448)) ([84e38db](https://github.com/hatayama/unity-cli-loop/commit/84e38dbc726110fbc37167b5dabd41608b59df7e))

## [3.0.0-beta.36](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.35...dispatcher-v3.0.0-beta.36) (2026-08-28)


### Features

* add set-code-optimization command and recommend permanent Debug for pause-point ([#2442](https://github.com/hatayama/unity-cli-loop/issues/2442)) ([3c54031](https://github.com/hatayama/unity-cli-loop/commit/3c540310a4e5a6a51ecccc5e04a47c6951a3352a))

## [3.0.0-beta.35](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.34...dispatcher-v3.0.0-beta.35) (2026-08-27)


### Bug Fixes

* skip dispatcher update download when already at the target version ([#2438](https://github.com/hatayama/unity-cli-loop/issues/2438)) ([cec5d09](https://github.com/hatayama/unity-cli-loop/commit/cec5d09584a51bcddcc4211d877508cc40b8ba29))

## [3.0.0-beta.34](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.33...dispatcher-v3.0.0-beta.34) (2026-08-27)


### Features

* retire the record-input CLI tool in favor of the Recordings window ([#2430](https://github.com/hatayama/unity-cli-loop/issues/2430)) ([54afc07](https://github.com/hatayama/unity-cli-loop/commit/54afc07bff66933dfcd23607e60317ce4c1419a5))

## [3.0.0-beta.33](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.32...dispatcher-v3.0.0-beta.33) (2026-08-25)


### ⚠ BREAKING CHANGES

* align CLI option names across first-party tools ([#2422](https://github.com/hatayama/unity-cli-loop/issues/2422))

### Features

* accept --file and --line on pause-point-status and await-pause-point ([#2411](https://github.com/hatayama/unity-cli-loop/issues/2411)) ([f650e36](https://github.com/hatayama/unity-cli-loop/commit/f650e36a6bcbd2880dba6ee385c59181895e756f))
* add a compact command-name listing to uloop list ([13f5d06](https://github.com/hatayama/unity-cli-loop/commit/13f5d06bba57bfb7f1afdb430dab4e2e7fd384d4))
* add a max-caller-frames option to enable-pause-point ([779c32b](https://github.com/hatayama/unity-cli-loop/commit/779c32b5af581bdca603be069f245711e4d4760d))
* add uloop hot-reload for live C# patching without domain reload ([#2254](https://github.com/hatayama/unity-cli-loop/issues/2254)) ([20e97f6](https://github.com/hatayama/unity-cli-loop/commit/20e97f69c6fd6f64095934a4eebdc18115329e6a))
* align CLI option names across first-party tools ([#2422](https://github.com/hatayama/unity-cli-loop/issues/2422)) ([2e4daff](https://github.com/hatayama/unity-cli-loop/commit/2e4daff6536a7febe88841d9c6104b2fbf8923ff))
* compile pending script changes before running tests ([5ab6770](https://github.com/hatayama/unity-cli-loop/commit/5ab6770185af6444b9074424fd88bb667cf8f67b))
* default hot-reload --files to sources changed since the last compile ([1c75e85](https://github.com/hatayama/unity-cli-loop/commit/1c75e85ade883553f691c91d116ed9501bb91835))
* default screenshot capture mode to rendering during Play Mode ([#2379](https://github.com/hatayama/unity-cli-loop/issues/2379)) ([393c11f](https://github.com/hatayama/unity-cli-loop/commit/393c11fa00e8cce3bb6e8058b0a06da56ec34233))
* focus a freshly launched Unity while waiting for readiness ([#2344](https://github.com/hatayama/unity-cli-loop/issues/2344)) ([a931e0d](https://github.com/hatayama/unity-cli-loop/commit/a931e0d3f8f98dcee1a40525a8fec3633c2760fb))
* gate pause-point hits with the hit-when condition ([#2405](https://github.com/hatayama/unity-cli-loop/issues/2405)) ([f5f4e33](https://github.com/hatayama/unity-cli-loop/commit/f5f4e3391e3a131b92d12288fbfe633b4a3c4914))
* include the running tool's elapsed time in single-flight busy rejections ([c5709bd](https://github.com/hatayama/unity-cli-loop/commit/c5709bd769e384249e2d0d4ea1235ad110e30b22))
* install skills into any directory with the new --output-dir option ([#2332](https://github.com/hatayama/unity-cli-loop/issues/2332)) ([f3c8051](https://github.com/hatayama/unity-cli-loop/commit/f3c8051de70e90c57ab03305649c44ca28c56d37))
* list all markers when pause-point-status omits --id ([#2413](https://github.com/hatayama/unity-cli-loop/issues/2413)) ([aee4288](https://github.com/hatayama/unity-cli-loop/commit/aee4288a2415efa148e335572442307402cb19e6))
* Remove the leftover uloop completion command ([50aa929](https://github.com/hatayama/unity-cli-loop/commit/50aa92925f963450e0becf6c113366b12ab2a42c))
* report interim compile-wait diagnostics so a stalled compile is visible before the timeout ([6d9c416](https://github.com/hatayama/unity-cli-loop/commit/6d9c4164f0797123084011054715fbc031a54dbb))


### Bug Fixes

* --trigger now rejects a leading uloop and shows the command to retry ([#2388](https://github.com/hatayama/unity-cli-loop/issues/2388)) ([9cefd73](https://github.com/hatayama/unity-cli-loop/commit/9cefd733403d6d5d74cde0cdbac511c1e80561bc))
* align clear-pause-point messages with ClearedCount for auto-disarmed and expired markers ([319159b](https://github.com/hatayama/unity-cli-loop/commit/319159bb6cffef137eef8c91eb97435a84b1fff2))
* describe CLI-only options in await-pause-point and pause-point-status help ([#2299](https://github.com/hatayama/unity-cli-loop/issues/2299)) ([7b237bd](https://github.com/hatayama/unity-cli-loop/commit/7b237bd2130c734733abbf3f920f6250a3885663))
* force TLS 1.2 and basic parsing for PowerShell installer downloads ([ea9e052](https://github.com/hatayama/unity-cli-loop/commit/ea9e0529a8b8565eac344a72582a1770c2c59d4e))
* Git Bash can install uloop from a zip without unzip ([cbcdc62](https://github.com/hatayama/unity-cli-loop/commit/cbcdc62b7fc1c1134522fb7f31cbb8c0c8863b9d))
* Windows uloop update can replace a running uloop.exe ([6857430](https://github.com/hatayama/unity-cli-loop/commit/6857430d1c33ab7e46c778eae318afdf2de53576))

## [3.0.0-beta.32](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.31...dispatcher-v3.0.0-beta.32) (2026-08-12)


### Features

* distribute uloop via a Homebrew tap ([#2157](https://github.com/hatayama/unity-cli-loop/issues/2157)) ([f882ea0](https://github.com/hatayama/unity-cli-loop/commit/f882ea0a9bc79cad97ae48dd4ea0ddc6e1a98dce))

## [3.0.0-beta.31] (2026-08-11)


## [3.0.0-beta.30](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.29...dispatcher-v3.0.0-beta.30) (2026-07-30)


### Bug Fixes

* terminal install works with curl alone by trusting the repository pin ([#2081](https://github.com/hatayama/unity-cli-loop/issues/2081)) ([58ec443](https://github.com/hatayama/unity-cli-loop/commit/58ec44303d69e1aa3e620f3c27b293218f4b5cc3))

## [3.0.0-beta.29](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.28...dispatcher-v3.0.0-beta.29) (2026-07-29)


### Features

* install the Unity package from the terminal with uloop package install ([#2078](https://github.com/hatayama/unity-cli-loop/issues/2078)) ([5e62a86](https://github.com/hatayama/unity-cli-loop/commit/5e62a86d880d05f15eb246ed5adbfa7ad5db74f7))

## [3.0.0-beta.28](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.27...dispatcher-v3.0.0-beta.28) (2026-07-29)


### Features

* check 3D physics hits with simulate-mouse-input --dry-run ([#2072](https://github.com/hatayama/unity-cli-loop/issues/2072)) ([24225f6](https://github.com/hatayama/unity-cli-loop/commit/24225f65af305fdb486ac80b36d3c0d01bb30d01))

## [3.0.0-beta.27](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.26...dispatcher-v3.0.0-beta.27) (2026-07-29)


### Features

* address pause point round 13-14 leftover improvements ([#2065](https://github.com/hatayama/unity-cli-loop/issues/2065)) ([ce21b4f](https://github.com/hatayama/unity-cli-loop/commit/ce21b4f5d1d3693cb9837288da15f4b43102fabe))

## [3.0.0-beta.26](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.25...dispatcher-v3.0.0-beta.26) (2026-07-29)


### Features

* address pause point round 13-14 verification feedback ([#2059](https://github.com/hatayama/unity-cli-loop/issues/2059)) ([62db173](https://github.com/hatayama/unity-cli-loop/commit/62db17327b467ab44250d1ca59f8b136129251ae))

## [3.0.0-beta.25](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.24...dispatcher-v3.0.0-beta.25) (2026-07-28)


### Features

* configurable compile wait timeout and working timeout recovery ([#2036](https://github.com/hatayama/unity-cli-loop/issues/2036)) ([fc867c3](https://github.com/hatayama/unity-cli-loop/commit/fc867c3914f979dbc8f31f26be4c5753f66e1c36))

## [3.0.0-beta.24](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.23...dispatcher-v3.0.0-beta.24) (2026-07-27)


### Bug Fixes

* V2 delegation notice now names the V2 CLI version that ran ([#2027](https://github.com/hatayama/unity-cli-loop/issues/2027)) ([5ed982d](https://github.com/hatayama/unity-cli-loop/commit/5ed982d6d523714355f33b7fe8856f773079afeb))

## [3.0.0-beta.23](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.22...dispatcher-v3.0.0-beta.23) (2026-07-27)


### Features

* Improve uloop CLI discoverability after E2E verification rounds ([#2023](https://github.com/hatayama/unity-cli-loop/issues/2023)) ([fcad6e7](https://github.com/hatayama/unity-cli-loop/commit/fcad6e7e7ff279ea160828caa938f9105d4a6c30))


### Bug Fixes

* Rename the global uloop launcher to dispatcher in user-facing text ([#1994](https://github.com/hatayama/unity-cli-loop/issues/1994)) ([3fdd062](https://github.com/hatayama/unity-cli-loop/commit/3fdd062d5821b746f2db142d7430a29a5a8ba9a0))
* Wait for V2 server readiness in uloop launch ([#2024](https://github.com/hatayama/unity-cli-loop/issues/2024)) ([0f076fc](https://github.com/hatayama/unity-cli-loop/commit/0f076fc20c9ede308cd6a9c3fb22302c58f85ef2))

## [3.0.0-beta.22](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.21...dispatcher-v3.0.0-beta.22) (2026-07-24)


### Bug Fixes

* Improve skills setup panel empty state and refresh scoping ([#1983](https://github.com/hatayama/unity-cli-loop/issues/1983)) ([caeb079](https://github.com/hatayama/unity-cli-loop/commit/caeb0799d78331162e1cae5c0572c8fc2ed6aa81))
* Tighten v3 migration skill guidance and align wizard prompt ([#1981](https://github.com/hatayama/unity-cli-loop/issues/1981)) ([7c8a4a2](https://github.com/hatayama/unity-cli-loop/commit/7c8a4a20a7cda853a9ddfdf966fa24c4ccfc5609))

## [3.0.0-beta.21](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.20...dispatcher-v3.0.0-beta.21) (2026-07-24)


### Features

* pause-point round 10 feedback improvements ([8b66c6a](https://github.com/hatayama/unity-cli-loop/commit/8b66c6a7e7066e82820b144b85a9b2d89ff071ed))


### Bug Fixes

* Windows CLI install robustness and uninstall message wording ([#1970](https://github.com/hatayama/unity-cli-loop/issues/1970)) ([40a12ad](https://github.com/hatayama/unity-cli-loop/commit/40a12add4b445293ee7a4820ae26c860068a2dac))

## [3.0.0-beta.20](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.0-beta.19...dispatcher-v3.0.0-beta.20) (2026-07-22)


### Features

* Round-8 pause-point usability improvements ([#1944](https://github.com/hatayama/unity-cli-loop/issues/1944)) ([1940cca](https://github.com/hatayama/unity-cli-loop/commit/1940cca1dea33377300cb06a13cb65ba9be85db9))

## [3.0.0-beta.19](https://github.com/hatayama/unity-cli-loop/releases/tag/dispatcher-v3.0.0-beta.19) (2026-07-21)

Version realignment: the dispatcher rejoins the 3.0.0-beta line. The 3.1.0-beta series existed only because an unintended dispatcher-v3.0.0 bootstrap release made release-please treat 3.0.0 as shipped. No functional changes.

## [3.1.0-beta.19](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.1.0-beta.18...dispatcher-v3.1.0-beta.19) (2026-07-20)


### Features

* dispatcher env override to use a locally built project runner ([#1863](https://github.com/hatayama/unity-cli-loop/issues/1863)) ([e6ec679](https://github.com/hatayama/unity-cli-loop/commit/e6ec67921fde80b5a8bdc5c3915618489d7f4615))


### Bug Fixes

* Launch no longer overclaims a crash for a stale UnityLockfile ([#1867](https://github.com/hatayama/unity-cli-loop/issues/1867)) ([7622db7](https://github.com/hatayama/unity-cli-loop/commit/7622db77b952243c1ca387ffca06af0fcf25217d))

## [3.1.0-beta.18](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.1.0-beta.17...dispatcher-v3.1.0-beta.18) (2026-07-20)


### Bug Fixes

* Runner-owned command flags and help no longer require a dispatcher release ([#1862](https://github.com/hatayama/unity-cli-loop/issues/1862)) ([96e75f8](https://github.com/hatayama/unity-cli-loop/commit/96e75f8a19c23e4c3d587e91efa3e52574716060))

## [3.1.0-beta.17](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.1.0-beta.16...dispatcher-v3.1.0-beta.17) (2026-07-19)


### Bug Fixes

* pause-pointの応答から重複情報を削減し、変数値だけを後から選んで取得できるように改善 ([#1857](https://github.com/hatayama/unity-cli-loop/issues/1857)) ([d507274](https://github.com/hatayama/unity-cli-loop/commit/d507274f14f61236aee30c3886139aa48c1a46d1))

## [3.1.0-beta.16](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.1.0-beta.15...dispatcher-v3.1.0-beta.16) (2026-07-19)


### Bug Fixes

* Detect V2 Unity projects with file:/embedded package references or ambiguous git cache generations ([#1847](https://github.com/hatayama/unity-cli-loop/issues/1847)) ([85509a8](https://github.com/hatayama/unity-cli-loop/commit/85509a8d5b50cf5354e0506332f5f335d012cfa4))

## [3.1.0-beta.15](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.1.0-beta.14...dispatcher-v3.1.0-beta.15) (2026-07-18)


### Bug Fixes

* align tool exit codes with response success ([#1824](https://github.com/hatayama/unity-cli-loop/issues/1824)) ([1cc123b](https://github.com/hatayama/unity-cli-loop/commit/1cc123b8d450b137ef489d55af0d13f9f587c16f))
* Launch V2 Unity projects without hanging ([#1826](https://github.com/hatayama/unity-cli-loop/issues/1826)) ([4ace5b4](https://github.com/hatayama/unity-cli-loop/commit/4ace5b49f3cbc7fa92ff7c3871f5064091c4b5cc))

## [3.1.0-beta.14](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.1.0-beta.13...dispatcher-v3.1.0-beta.14) (2026-07-17)


### Features

* Delegate V2 projects through the V3 dispatcher ([#1807](https://github.com/hatayama/unity-cli-loop/issues/1807)) ([3882b19](https://github.com/hatayama/unity-cli-loop/commit/3882b1913184dcbce0f94f6e5b6cf806b7405eb1))


### Bug Fixes

* make Windows v3 workflows reliable ([#1818](https://github.com/hatayama/unity-cli-loop/issues/1818)) ([21eae0a](https://github.com/hatayama/unity-cli-loop/commit/21eae0a96af05355cbf57eb3ab98dd7388fc7b2a))
* Project commands now work from paths containing glob characters ([#1810](https://github.com/hatayama/unity-cli-loop/issues/1810)) ([c56843a](https://github.com/hatayama/unity-cli-loop/commit/c56843a11e7e794d27c378dbaddeac90bf87be75))
* V2 delegation now recognizes SCP Git dependencies with path queries ([#1811](https://github.com/hatayama/unity-cli-loop/issues/1811)) ([b42f918](https://github.com/hatayama/unity-cli-loop/commit/b42f91857c51556102a35514da89c305d3feb1bf))

## [3.1.0-beta.13](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.1-beta.13...dispatcher-v3.1.0-beta.13) (2026-07-15)


### Features

* add pause-point watch expressions ([#1733](https://github.com/hatayama/unity-cli-loop/issues/1733)) ([875835f](https://github.com/hatayama/unity-cli-loop/commit/875835fc71deb70ccdc0bea0d67582a0d42a2f80))
* Improve CLI guidance for tool enums, busy errors, and dynamic-code diagnostics ([7318b9d](https://github.com/hatayama/unity-cli-loop/commit/7318b9d1dcead7b52241e88adab7538b47a5f95a))
* Pause point wait command is now await-pause-point ([#1698](https://github.com/hatayama/unity-cli-loop/issues/1698)) ([f1d0a9d](https://github.com/hatayama/unity-cli-loop/commit/f1d0a9d1c6c72a8699a2468e68a05262a50642dc))


### Bug Fixes

* bound Go external OS commands and propagate Ctrl+C cancellation ([#1738](https://github.com/hatayama/unity-cli-loop/issues/1738)) ([ddb0581](https://github.com/hatayama/unity-cli-loop/commit/ddb058124507d0207d371bcf988ef6449b0c0b66))
* compile-consistency — external scene hold, compile wait/TTL align, API Update guidance ([#1760](https://github.com/hatayama/unity-cli-loop/issues/1760)) ([247cb0c](https://github.com/hatayama/unity-cli-loop/commit/247cb0c62a6a87fd56dba0126334fb5061d4d081))
* Harden CLI distribution and Unity IPC security ([#1794](https://github.com/hatayama/unity-cli-loop/issues/1794)) ([b5ca16b](https://github.com/hatayama/unity-cli-loop/commit/b5ca16b34fc8359466183c0cac30f2d77e862212))
* Harden IPC contracts, empty RPC errors, and Settings async UI ([#1778](https://github.com/hatayama/unity-cli-loop/issues/1778)) ([0dffc75](https://github.com/hatayama/unity-cli-loop/commit/0dffc753bec575a29252430a59540ad3e0812848))
* Harden V2-to-V3 third-party migration for safe apply and reliable scans ([#1710](https://github.com/hatayama/unity-cli-loop/issues/1710)) ([b4fbb0d](https://github.com/hatayama/unity-cli-loop/commit/b4fbb0db6b8837e4036637e11684bc08da427935))

## [3.0.1-beta.13](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.1-beta.12...dispatcher-v3.0.1-beta.13) (2026-07-11)


### Bug Fixes

* Avoid misclassifying copied project error text ([0e821cc](https://github.com/hatayama/unity-cli-loop/commit/0e821cc94b12a03a0131ca64f76da2385432b23e))
* Busy responses now use the standard error envelope ([0aa2e7e](https://github.com/hatayama/unity-cli-loop/commit/0aa2e7e3355cb5e808c460cd3707a5f2ec0e7525))
* Dispatcher requires the project runner pin instead of parsing CliConstants ([#1501](https://github.com/hatayama/unity-cli-loop/issues/1501)) ([2012eb8](https://github.com/hatayama/unity-cli-loop/commit/2012eb882cf5bef0de72a49d2cda08f12586daf3))
* **dispatcher:** Verify dispatcher archive attestation before installer runs ([#1673](https://github.com/hatayama/unity-cli-loop/issues/1673)) ([1a49534](https://github.com/hatayama/unity-cli-loop/commit/1a4953464d7fa335ed73c2b71ef1e88485769ff6))
* **dispatcher:** Verify installer attestation before running self-update ([#1671](https://github.com/hatayama/unity-cli-loop/issues/1671)) ([109bc9f](https://github.com/hatayama/unity-cli-loop/commit/109bc9f2202b9adea817276624a6f5be77133c34))
* **dispatcher:** Verify project runner archive attestation before extraction ([#1672](https://github.com/hatayama/unity-cli-loop/issues/1672)) ([fa3313d](https://github.com/hatayama/unity-cli-loop/commit/fa3313d64ba98455015afc12acd6b407b4c52681))
* **install:** Validate ULOOP_VERSION shape before use ([#1675](https://github.com/hatayama/unity-cli-loop/issues/1675)) ([c09b683](https://github.com/hatayama/unity-cli-loop/commit/c09b683bd10c25626a72d7cac14db235f744aa1c))
* Local skill packages no longer include stale cached skills ([#1615](https://github.com/hatayama/unity-cli-loop/issues/1615)) ([9388b91](https://github.com/hatayama/unity-cli-loop/commit/9388b91801ea4ed76aa5c1251861a0efb9210d6c))
* Native uninstall now removes stale shell PATH setup ([#1498](https://github.com/hatayama/unity-cli-loop/issues/1498)) ([3450a1e](https://github.com/hatayama/unity-cli-loop/commit/3450a1ebe22d5a65f0c75d5ef1cd2a761ec3e5d5))
* Prefer synced tool definitions for help and completion ([603d178](https://github.com/hatayama/unity-cli-loop/commit/603d1783b0f67116d2f14091a02016d5e343d9f6))
* Read minimum version requirements from the project runner pin ([#1506](https://github.com/hatayama/unity-cli-loop/issues/1506)) ([18bf780](https://github.com/hatayama/unity-cli-loop/commit/18bf780dc31f6105902c383186e97c4aec7c8772))
* Reject malformed dispatcher minimum versions consistently ([aa8a2fc](https://github.com/hatayama/unity-cli-loop/commit/aa8a2fc0dd2bff438dbc1c58ed7216333342aa81))
* Remove the dispatcher contract integer generation ([#1504](https://github.com/hatayama/unity-cli-loop/issues/1504)) ([d2ddbce](https://github.com/hatayama/unity-cli-loop/commit/d2ddbce87d9bd68b53a1e202e9ce43996b61acf6))
* Retry incomplete project runner downloads ([3c8cea0](https://github.com/hatayama/unity-cli-loop/commit/3c8cea0e80c514e7f72684d11b13dbb38ffc600b))
* Shared IPC clients assign request IDs safely ([b7d6acc](https://github.com/hatayama/unity-cli-loop/commit/b7d6acc161083a3bd62ce9e66ba0d4f00937f05f))
* Shrink the project runner pin schema to two fields ([#1503](https://github.com/hatayama/unity-cli-loop/issues/1503)) ([adceaed](https://github.com/hatayama/unity-cli-loop/commit/adceaedd78a8b7f608f5c2179e1ee799fdecb136))
* Verify dispatcher self-update installers ([0d132e2](https://github.com/hatayama/unity-cli-loop/commit/0d132e2b6cb10691cbd7ed9299b11bc4da8d03b1))

## [3.0.1-beta.12](https://github.com/hatayama/unity-cli-loop/compare/dispatcher-v3.0.1-beta.11...dispatcher-v3.0.1-beta.12) (2026-07-03)


### Bug Fixes

* Correct misleading dispatcher error and help messages ([#1464](https://github.com/hatayama/unity-cli-loop/issues/1464)) ([439835f](https://github.com/hatayama/unity-cli-loop/commit/439835f716efe48d7fed2376e6f4eff0f1bdeba7))

## [3.0.1-beta.11](https://github.com/hatayama/unity-cli-loop/releases/tag/dispatcher-v3.0.1-beta.11) (2026-07-03)

Initial changelog entry. Earlier dispatcher releases predate Release Please management; see the GitHub releases page for their notes.
