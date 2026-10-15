# AUDIT-0218 - Record CubeCraft release routing blocker

- Commit subject: `docs: record CubeCraft release routing blocker`
- Change type: documentation
- Runtime impact: none. No protocol or forwarding code changed.

## Intent and scope

The maintainer requested a v0.4.0 release. The release check used the clean runtime revision `0b597914c7b20eb988a2d7840bb0e91a4d908722`. Five required retail-client server reports passed. The CubeCraft selector did not return a usable destination, so the release could not be published. The research note records the observed response shape, the resolver boundary, and what remains unknown without retaining service values or credentials.

## Evidence and provenance

The current `experience:CubeCraft` route failed twice before a downstream listener opened. A temporary diagnostic build inspected only field names and value shapes, and its source changes were restored. `internal/experience/target.go`, `docs/research/experience-routing.md`, Mojang's NetherNet onboarding guide inspected on 2026-10-15, and the installed go-nethernet v1.0.24 dialer were reviewed. No CubeCraft capture exists for these attempts. No server connection, private message, or packet mutation was performed.

## Validation and limits

The five generated retail reports for The Hive, Galaxite, Lifeboat, Mineville Zeqa, and Enchanted all returned `pass` for the same runtime revision. `tools/check-release-readiness.ps1` reported CubeCraft `missing` and exited with status 1. The temporary diagnostic change was removed and the worktree returned clean before this documentation change. No runtime fix has been demonstrated, and the required CubeCraft live session remains unobserved. Release publication remains blocked until the six-server gate passes.

The canonical `tools/quality.ps1 -LowMemory` gate and `git diff --check` passed after this documentation change. The vulnerability scan found zero reachable vulnerabilities and one imported or required advisory outside called code. The earlier runtime candidate's exact-revision CI completed successfully on Linux and Windows; this documentation commit still requires its own exact-revision CI after push.
