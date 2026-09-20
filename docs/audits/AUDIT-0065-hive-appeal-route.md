# AUDIT-0065 - Document The Hive appeal route

- Commit subject: `docs: document The Hive appeal route`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record the official route for disputing the current The Hive game ban and distinguish it from a future VPN or proxy exemption request.

## Scope

Add official Hive support links, required appeal fields, the rule that appeal decisions are final, the fallback when a Ban ID is missing, and the instruction to keep connection identifiers private. No source, protocol model, capture, credential, or runtime behavior changes.

## Evidence and provenance

The note cites Hive's official Punishments and Appeals page, Punishment ID guide, Player Support form, VPN policy, and a 2026 Hive forum staff reply describing the `Game Ban` fallback when the ID is not displayed. The screenshot itself is not copied into the repository and its identifiers are not recorded.

## Validation

- Automated checks: `git diff --check`, the existing quality gate, and targeted documentation searches.
- Live validation: `not_observed`. Documentation does not authorize another Hive connection.
- Capture or report: existing local captures remain unchanged and sensitive.

## Limitations and follow-up

Hive's support team alone can decide the appeal. The project cannot infer the detector or guarantee that the punishment will be reversed. No runtime change is proposed.

No runtime impact.
