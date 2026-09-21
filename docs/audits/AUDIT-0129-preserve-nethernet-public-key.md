# Audit 0129 - Preserve the NetherNet transport public key

- Intent: keep the optional authenticated transport capability visible through BedrockDebugProxy's capture wrapper.
- Scope: `internal/bedrock/network.go`, its regression test, the ecosystem research note, fix history, and the Unreleased changelog.
- Behavior: `observedConn.PublicKey` forwards the underlying transport key when available and returns nil otherwise. Gophertunnel can therefore retain its existing transport-to-login identity check for NetherNet connections without any packet or token rewrite.
- Evidence and provenance: gophertunnel's unexported `publicKeyConn` interface is a method-shape boundary in `third_party/gophertunnel/minecraft/conn.go`. The selected go-nethernet module exposes the matching method on its connection. No source was copied from that dependency.
- Automated validation: `go test ./internal/bedrock -count=1`, the project low-memory quality gate, and `git diff --check` are required before commit.
- Live validation: not run. The user had quit the local session and no public or external server was contacted.
- Known limitations: the local test verifies capability preservation but not a live NetherNet handshake or remote identity result.
- Follow-up: exercise a controlled NetherNet loopback or owner-controlled endpoint when available and inspect the capture's transport connection and login events separately.
