# AUDIT-0110 - Align resource-pack reader documentation

- Commit subject: `Align resource-pack reader documentation`
- Change type: `documentation`
- Runtime impact: `none`

## Intent

Keep the resource-pack research note aligned with the bounded, context-aware helper used by the current URL adapter.

## Scope

Replace the stale `resource.ReadURL` reference in `docs/resource-packs.md` with `resource.ReadURLLimitedContext`.
No code, capture schema, protocol behavior, or validation status changes.

## Evidence and provenance

`internal/proxy/resource_pack_url.go` calls the context-aware helper introduced by the preceding URL resource-pack
cancellation change. The documentation-only correction reflects the current source path and does not add a new
protocol claim.

## Automated validation

- `git diff --check`
- The next full quality gate remains required for any later runtime change.

## Live validation status

Not applicable. No runtime behavior changed and no server was contacted.

## Limitations and follow-up

The note still records that real-client URL and mixed-offer resource-pack validation is pending.
