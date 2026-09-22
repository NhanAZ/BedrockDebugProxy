# Audit AUDIT-0172

## Intent

Fix the Galaxite release-gate decode error for a valid observed `ClientCacheBlobStatus` list containing 1,302 hashes without weakening generic protocol slice limits.

## Scope

Updated the vendored gophertunnel protocol helper and `ClientCacheBlobStatus` packet to support a packet-specific 16,384-entry bound. Added a root protocol regression test, provenance documentation, fix history, research status, and changelog entry. Raw forwarding, capture storage, and packet mutation behavior are unchanged.

## Evidence and provenance

The change is based on capture `fe9d3e3b-aa80-4976-a6ec-31db99dd1db1`, where sequence 51,524 reported 1,302 hashes against the generic 1,024-entry bound. The source comparison and immutable upstream references are recorded in `docs/research/galaxite-client-cache-blob-status-limit.md`. The vendored patch note records the targeted compatibility change and retains the upstream MIT provenance.

## Automated validation

- `gofmt` passed for changed Go files.
- `go test ./internal/bedrock` passed, including `TestClientCacheBlobStatusAllowsObservedHashCount`.
- Full quality gate and exact-revision build remain required before live validation.

## Live validation status

The previous Galaxite report remains `fail` for runtime revision `93435c5119e2899dc1958d8cec0fa63447bd611c`. A fresh stamped build and Galaxite session are required to verify that the decode error is gone.

## Known limitations and follow-up

The 16,384-entry bound is a conservative memory guard chosen to cover the observed packet, not a claim that it is the protocol's universal maximum. A future protocol source or capture may justify a different bound. The release matrix must be rerun from the new runtime revision.
