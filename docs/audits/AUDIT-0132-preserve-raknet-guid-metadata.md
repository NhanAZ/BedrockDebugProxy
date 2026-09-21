# Audit 0132 - Preserve RakNet client GUID metadata

- Intent: retain a transport identity value already exposed by gophertunnel so local parity analysis can compare the RakNet connection without inspecting raw datagrams.
- Scope: `internal/proxy/proxy.go`, the transport metadata assertion in the proxy integration test, capture-format and capture-observability research, fix history, and the Unreleased changelog.
- Behavior: `transport.payload` events include `raknet_client_guid` when `Conn.RakNetClientGUID` reports a non-zero value. Other transports and early unavailable states omit it.
- Evidence and provenance: gophertunnel delegates `Conn.RakNetClientGUID` to the optional transport `ClientGUID` method, and BedrockDebugProxy already uses the same value when following `Transfer`. The value is retained as sensitive local capture metadata and is not sent or transformed.
- Automated validation: proxy loopback integration test, project low-memory quality gate, and `git diff --check` are required before commit. The integration test verifies the dialled upstream channel; listener-side transports may omit the optional value.
- Live validation: not run. The user quit the local session and no public or external server was contacted.
- Known limitations: the field does not expose raw UDP, RakNet acknowledgements, fragmentation, retransmissions, or a non-RakNet transport identity.
- Follow-up: keep transport identity fields bounded and explicitly classified before adding any lower-layer capture.
