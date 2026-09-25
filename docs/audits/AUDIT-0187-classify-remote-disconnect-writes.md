# Audit 0187 - Classify remote disconnect writes

- Commit subject: `Classify remote disconnect writes as peer termination`
- Change type: `connection termination classification fix`
- Runtime impact: `yes`

## Intent

Avoid reporting writes to an already-closed upstream connection as local bridge write failures when gophertunnel reports the peer's explicit Bedrock `Disconnect` packet as a typed `minecraft.DisconnectError`.

## Scope

- Treat wrapped `minecraft.DisconnectError` as a peer-termination error in coordinated bridge shutdown.
- Omit the downstream `RequestChunkRadius` callback error when its write failed because of the same typed remote disconnect.
- Add focused regression coverage and update the changelog and fix history.

## Evidence and provenance

The local Enchanted captures [captures/session-20260924T152209Z](https://github.com/BedrockDebugProxy-Captures/session-20260924T152209Z/releases/tag/capture-archive) at `bfe0741fa4cc9fb74199bde0624ce38c16d4af76` and [captures/session-20260923T062327Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T062327Z/releases/tag/capture-archive) at `bfc19102a312cbe5b0a18c9432c37afcbb446b24` both recorded upstream packet ID 5, the protocol `Disconnect` ID, followed by upstream closure and later bridge write errors. Source review of the in-tree gophertunnel `Conn.receive` shows that it closes the connection with a typed `DisconnectError`; `Conn.closeErr` wraps the close cause for subsequent writes. The private packet message text was not inspected or added to this record.

## Validation

- Focused regression checks: `go test ./internal/proxy -run 'Test(ExpectedShutdownErrorRecognizesRemoteDisconnect|FinishForwardSuppressesFirstRemoteDisconnectWriteError|ExpectedShutdownErrorRequiresCoordinatedShutdown|FinishForwardDoesNotRecordExpectedShutdownWriteError|FinishForwardCoordinatesRacingShutdownErrors|FinishForwardSuppressesFirstPeerCloseWriteError)$'` passed.
- Project quality gate: `tools/quality.ps1 -LowMemory` passed formatting, PowerShell and viewer JavaScript syntax, module integrity, third-party license inventory, the full Go test suite, build, static analysis (0 issues), vulnerability scan (0 reachable vulnerabilities; 4 in imported packages and 2 in required modules not called by this code), tracked capture artifact checks, and whitespace validation.
- Live validation: pending. A stamped build must complete a fresh Enchanted session and The Hive baseline. The upstream `Disconnect` remains visible evidence, and the operator must still judge whether the session ended normally.
- CI: pending for the commit containing this record.

## Limitations and follow-up

This change only corrects classification of downstream writes after a typed remote disconnect. It does not suppress or rewrite the `Disconnect`, prevent the server from closing the connection, or convert a failed manual validation check into a pass. The six-server release matrix remains required on the validated runtime revision.
