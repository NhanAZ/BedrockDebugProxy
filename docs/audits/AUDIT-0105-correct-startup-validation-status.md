# Audit 0105 - Correct startup validation status

- Intent: remove a stale limitation sentence that claimed direct vanilla-client validation was still pending after the controlled real-client loopback was recorded.
- Scope: documentation only. Updated the startup parity limitation wording.
- Affected interfaces: `docs/research/start-game-parity.md` only. No Go code, capture schema, protocol behavior, binary output, or runtime configuration changed.
- Evidence and provenance: the owner-controlled real-client capture `captures/session-20260921T122228Z` and the post-fix loopback section in the same research note.
- Automated validation: `tools/quality.ps1 -LowMemory` and `git diff --check` pass before commit.
- Local validation: no new runtime test was required. The referenced capture already verified the real-client loopback and startup payload hashes.
- Live validation: no public server was contacted. Public-server validation remains pending.
- Known limitations: this wording correction does not establish authenticated public-server behavior or raw UDP/RakNet parity.
- Follow-up: keep limitation text synchronized with the evidence sections when new validation runs complete.
