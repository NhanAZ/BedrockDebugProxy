# Audit 0130 - Capture transport status-ping responses

- Intent: preserve status-ping response evidence that is already observable at the selected `minecraft.Network` boundary.
- Scope: `internal/bedrock/network.go`, network tests, capture-format and ecosystem research docs, fix history, and the Unreleased changelog.
- Behavior: `Network.PingContext` records a `transport.ping` event with the requested address, byte count, duration, errors, and any response bytes. Successful or partial responses are stored as a raw blob with representation `bedrock_transport_ping`.
- Evidence and provenance: gophertunnel's `Dialer.prepareDialAddress` consumes the response to apply a pong-advertised port, and transfer probing calls the same network operation. The request's raw UDP datagram is not exposed by this interface and is not claimed.
- Automated validation: network unit tests, the project low-memory quality gate, and `git diff --check` are required before commit.
- Live validation: not run. The user had quit the local session and no public or external server was contacted.
- Known limitations: ping requests and RakNet framing, loss, acknowledgements, and retransmissions remain outside the observation boundary.
- Follow-up: add a lower transport adapter only if raw UDP or PCAP evidence becomes an explicit supported capture requirement.
