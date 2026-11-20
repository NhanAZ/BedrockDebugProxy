# AUDIT-0272 - Restore full-session release validation

- Commit subject: `Restore full-session release validation docs`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Align the release workflow with the current requirement to retain bidirectional decoded traffic for 30 seconds after spawn while keeping the client stationary and out of minigames.

## Scope

Record the current release gate in decision 0019, mark decision 0018 as historical, and update the changelog, release workflow, validation guide, server targets, and fix history. These changes do not alter the Golem client or capture format.

## Evidence and provenance

The exact `0.6.1` candidate is stamped from runtime commit `9a30db76d47111344230c167824b9c57a56bad04`. All six release reports record `cli_session=pass`, `authenticated_upstream=pass`, and `post_spawn_stability=pass`, with verified captures and completed artifact folders. The first 30 seconds after each spawn contain bidirectional decoded traffic:

| Server | Post-spawn seconds | Client to server decoded packets | Server to client decoded packets | Encrypted packs decrypted |
| --- | ---: | ---: | ---: | ---: |
| The Hive | 30.2 | 890 | 3,752 | 23 / 23 |
| Galaxite | 30.5 | 600 | 2,729 | 20 / 20 |
| Lifeboat | 30.2 | 602 | 3,646 | 13 / 13 |
| Mineville Zeqa | 30.2 | 602 | 334 | 0 |
| Enchanted | 30.4 | 599 | 3 | 0 |
| CubeCraft | 30.2 | 600 | 4,559 | 5 / 5 |

The corrected NetherNet behavior also passed a live OneBlockOnline packs-only exchange. Its verified capture [session-20261120T210631Z](https://github.com/BedrockDebugProxy-Captures/session-20261120T210631Z/releases/tag/capture-archive) contains three imported packs and no unsupported `transport.ping` diagnostic. The earlier failing baseline is archived at [session-20261120T195739Z](https://github.com/BedrockDebugProxy-Captures/session-20261120T195739Z/releases/tag/capture-archive).

## Validation

- Automated checks: the runtime commit passed `tools/quality.ps1`. The exact stamped `0.6.1` binary passed the focused BedrockPackArchive producer-verification fixture in paired consumer commit [`902cb09`](https://github.com/NhanAZ-Tools/BedrockPackArchive/commit/902cb0993cc6d573a73814fd753c0e1daa59fca7). The six-report `tools/check-release-readiness.ps1` gate passed for runtime revision `9a30db76d47111344230c167824b9c57a56bad04`.
- Live validation: pass for The Hive, Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and CubeCraft. Enchanted's Hub form remained unanswered. CubeCraft stopped at its lobby.
- Capture or report: sanitized reports are under `validation/local/9a30db76d47111344230c167824b9c57a56bad04/`. Raw captures remain in ignored local storage until release, private-handoff synchronization, and cleanup gates finish.

## Limitations and follow-up

The six reports establish the authenticated protocol flow and post-spawn stability observed by the Go client. They do not establish retail rendering or gameplay. Run the release-readiness checker in documentation-only mode after the next documentation commit and before publishing. The paired BedrockPackArchive producer compatibility update is complete. The fresh 14-Experience snapshot batch is still in progress.
