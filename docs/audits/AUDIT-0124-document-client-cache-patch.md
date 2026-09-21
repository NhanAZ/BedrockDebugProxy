# Audit AUDIT-0124 - Document client-cache patch provenance

- Commit subject: `Document client-cache patch provenance`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Keep the vendored gophertunnel patch inventory aligned with the project-specific
listener and dialer hooks introduced by the client-cache parity fix.

## Scope

- Document the opt-in deferred resource-pack offer mode in
  `third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md`.
- State that the mode passes the observed downstream capability to the paired
  dialer's existing `EnableClientCache` field without changing default listener
  behavior or forwarded packet bytes.

## Evidence and provenance

The implementation is documented in
`docs/decisions/0011-forward-client-cache-capability.md` and
`docs/research/client-cache-parity.md`. No source code is copied or changed by
this documentation-only update.

## Validation

- Automated validation: documentation diff review and `git diff --check`.
- Live validation: not applicable. This commit has no runtime impact.

## Limitations and follow-up

The runtime behavior and its focused automated coverage remain in commit
`01fcb6e`. A controlled local cache-blob and miss-response fixture is still
needed to validate the later cache data path.
