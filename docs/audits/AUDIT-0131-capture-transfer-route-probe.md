# Audit 0131 - Capture transfer route-probe responses

- Intent: retain status-ping response evidence from the pre-reconnect transfer route probe without changing its transport settings or routing role.
- Scope: `internal/proxy/proxy.go`, proxy regression tests, transfer research, fix history, and the Unreleased changelog.
- Behavior: route probes now call `bedrock.Network.PingContext` with the existing RakNet local address and client GUID. The route outcome remains `transfer.route_probe`; returned bytes and operation errors are stored as `transport.ping` on the next upstream hop.
- Evidence and provenance: the transfer path already records the previous upstream source address and RakNet client GUID. The route probe uses those values before client reconnect, while the subsequent actual dial continues to perform its own bounded ping. This change does not claim raw UDP or RakNet acknowledgement evidence.
- Automated validation: proxy route-probe regression test, project low-memory quality gate, and `git diff --check` are required before commit.
- Live validation: not run. The user quit the local session and no public or external server was contacted.
- Known limitations: raw ping requests, RakNet framing, loss, acknowledgements, and retransmissions remain outside the observation boundary.
- Follow-up: add a lower transport adapter only if raw UDP or PCAP evidence becomes an explicit supported capture requirement.
