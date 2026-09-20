# AUDIT-0072 - Document gophertunnel chunk findings

- Commit subject: `docs: record gophertunnel chunk compatibility findings`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record the official gophertunnel reports about Hive subchunk offsets and invisible blocks, and connect them to the proxy's existing raw-forwarding evidence.

## Scope

Update `docs/research/cubecraft-gophertunnel-discord-audit.md` only. The change adds source-backed context and does not alter packet handling, capture behavior, or server targets.

## Evidence and provenance

- Sandertv/gophertunnel issue [#350](https://github.com/Sandertv/gophertunnel/issues/350) reports a Hive subchunk offset.
- Sandertv/gophertunnel issue [#273](https://github.com/Sandertv/gophertunnel/issues/273) reports invisible blocks after gophertunnel chunk handling.
- The existing local regression test confirms unchanged raw `LevelChunk` payload hashes across the same-protocol bridge.

## Validation

- Automated: `git diff --check`.
- Runtime tests are unchanged. No public server, account token, or live capture was used.

## Limitations and follow-up

The notes do not identify any public server's anti-cheat rule. A local raw-forwarding result cannot establish public Hive compatibility or replace authorized live validation.
