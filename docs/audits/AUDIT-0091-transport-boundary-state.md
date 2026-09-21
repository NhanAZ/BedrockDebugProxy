# Audit AUDIT-0091 - Classify transport boundary state

- Commit subject: `Classify transport payload state`
- Change type: `capture observability`
- Runtime impact: `no wire behavior change`

## Intent

Make each transport payload event explain which compression and encryption state was observable at the associated Bedrock connection boundary without replacing the raw application-batch bytes.

## Scope

- Add a bounded metadata callback to the observed transport wrapper.
- Expose whether a vendored gophertunnel connection has enabled or disabled packet encryption.
- Record compression negotiation state, threshold, and algorithm when the connection has exposed NetworkSettings.
- Keep early payloads lossless when the connection state is not yet available and retain the existing capture schema version because the event data additions are optional and backward-compatible.

## Evidence and provenance

- The transport wrapper records application-batch bytes above RakNet reassembly, as documented in `docs/research/bedrock-ecosystem.md`.
- Gophertunnel enables compression from NetworkSettings and encryption after the Bedrock handshake in `minecraft/conn.go`.
- No packet bytes, forwarding decisions, compression policy, encryption policy, authentication behavior, or external service interaction are changed.

## Validation

- `go test ./...` in `third_party/gophertunnel` passed.
- `go test ./...` in the root module passed.
- The transport wrapper unit test verifies metadata fields and raw blob retention.
- The root proxy loopback integration test observes derived state on both upstream and downstream transport channels.
- `tools/quality.ps1` and `git diff --check` are required before commit.
- Live vanilla-client validation: `not_observed`.

## Limitations and follow-up

Metadata is unavailable for payloads recorded before the corresponding connection reference is established. The fields do not expose raw UDP datagrams, RakNet acknowledgements, fragmentation, or retransmission state. Compression settings remain observational; this change does not normalize either side's policy.
