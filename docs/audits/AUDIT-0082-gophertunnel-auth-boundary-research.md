# Audit AUDIT-0082 - Record authenticated key-binding research

- Commit subject: `Record authenticated key-binding research`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record an upstream interoperability report that explains why a terminating proxy cannot safely copy authenticated client metadata into a separately encrypted upstream session without a protocol-supported key-binding design.

## Scope

- Updated `docs/research/local-identity-parity.md` with gophertunnel issue #407 and its protocol scope.
- No runtime code, packet handling, authentication behavior, capture format, or CLI behavior changed.

## Evidence and provenance

- Sandertv/gophertunnel issue #407, reviewed 2026-09-21: <https://github.com/Sandertv/gophertunnel/issues/407>.
- The report covers BDS protocol 944 and is explicitly treated as corroborating interoperability evidence, not proof about The Hive or protocol 2193.
- No source code was copied and no credentials, captures, or private payloads were added.

## Validation

- Automated checks: `git diff --check` and the project quality gate.
- Live validation: `not_observed`.
- Capture or report: `none`.

## Limitations and follow-up

This is a documentation-only research update. No runtime impact. Authenticated parity still requires an authorized owner-controlled test and a protocol-supported identity design.
