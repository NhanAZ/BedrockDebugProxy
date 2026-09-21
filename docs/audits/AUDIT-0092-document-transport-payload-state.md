# Audit AUDIT-0092 - Document transport payload state

- Commit subject: `Document transport payload state`
- Change type: `documentation`
- Runtime impact: `none`

## Intent

Document the optional derived fields added to `transport.payload` events so capture readers can distinguish application-batch evidence from connection-state metadata.

## Scope

- Update `docs/capture-format.md` with the field names, allowed state values, and lower-layer boundary.
- Add this audit record for the documentation-only follow-up to `ceba2a5`.
- No source, capture schema constant, packet path, or viewer behavior changes.

## Evidence and validation

- The documented fields and state values match `internal/proxy/transportMetadata` and the transport wrapper tests.
- `git diff --check` and the project quality gate are required before commit.
- Live validation: `not_applicable`; this commit changes documentation only.

## Limitations

The documentation does not expand the transport observation boundary. Existing captures remain immutable and older events do not gain the new optional fields.
