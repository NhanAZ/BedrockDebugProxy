# AUDIT-0066 - Clarify The Hive InitialConnection-55 diagnostic

- Commit subject: `docs: clarify The Hive connection diagnostic`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Document the distinction between the client-side `InitialConnection-55` label in the ban screenshot and the authoritative Hive punishment response.

## Scope

Add the observed diagnostic fields and the limits of public documentation to the Hive research note. No source, protocol model, capture, credential, or runtime behavior changes.

## Evidence and provenance

The screenshot shows `InitialConnection-55` alongside `You are banned from The Hive`, `Unfair Advantage`, and `RakNet:2193`. Searches found public reports of other `InitialConnection` values but no authoritative mapping for `-55`. Connection, world-session, and client identifiers are not copied into the repository.

## Validation

- Automated checks: `git diff --check`, the existing quality gate, and targeted documentation searches.
- Live validation: `not_observed`. Documentation does not authorize another Hive connection.
- Capture or report: existing Hive captures are archived in private repositories and remain sensitive.

## Limitations and follow-up

The exact client implementation behind `InitialConnection-55` is unknown. The ban reason must be appealed as a Hive punishment, not treated as a standalone transport failure. No runtime change is proposed.

No runtime impact.
