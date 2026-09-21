# Audit 0133 - Record transport identity boundaries

- Intent: preserve optional transport identity capabilities at connection lifecycle boundaries for local parity analysis.
- Scope: `internal/bedrock/network.go`, network regression tests, capture-format and capture-observability research, fix history, and the Unreleased changelog.
- Behavior: connection lifecycle events retain a non-zero RakNet client GUID and public-key capability and presence booleans when the underlying connection exposes those methods. Public-key material is not serialized.
- Evidence and provenance: the gophertunnel transport contract and the observing wrapper expose the optional `ClientGUID` and `PublicKey` methods at the `net.Conn` boundary. The fields remain local sensitive metadata and do not affect forwarding.
- Automated validation: bedrock network regression tests, project low-memory quality gate, and `git diff --check` are required before commit.
- Live validation: not run. The user quit the local session and no public or external server was contacted.
- Known limitations: no raw UDP, RakNet acknowledgement, fragmentation, retransmission, or cryptographic key material is captured by these lifecycle fields.
- Follow-up: keep identity metadata bounded and capability-scoped; do not infer proxy attribution from a single field.
