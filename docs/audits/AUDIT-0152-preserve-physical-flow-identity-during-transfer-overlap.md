# Audit 0152 - Preserve physical flow identity during transfer overlap

## Intent

Keep packet provenance tied to the physical transport flow that produced it when a transfer reconnect overlaps the old client connection.

## Scope

- Updated `internal/bedrock/observer.go` to retain multiple physical flows per channel and select connection identity by source and destination addresses.
- Updated `internal/bedrock/network.go` to record connection ID and hop when each observed transport opens.
- Added overlapping-flow observer coverage and extended the transfer loopback regression with a post-Transfer client packet.
- Updated the capture observability research note, fix history, and changelog.
- No packet bytes, forwarding policy, transfer rewrite, capture schema, or protocol definitions changed.

## Evidence and provenance

- The transfer runner creates pending hop state before the old downstream client flow has necessarily closed.
- The focused regression reproduced the provenance issue. A client `RequestChunkRadius` sent after the old hop's `Transfer` was observed was captured on the downstream channel with `downstream-2` before the fix, even though its physical flow was the old connection. The corrected capture assertion requires `downstream-1`.
- The observer unit test uses two synthetic local UDP flows that overlap and verifies each raw event receives the identity recorded at transport open. No credentials, captures, or external endpoints are involved.

## Automated validation

- `go test ./internal/bedrock ./internal/proxy -run 'TestObserverUsesPhysicalFlowIdentityWhenFlowsOverlap|TestRunnerFollowsTransferAndAcceptsNextHop' -count=1 -v` passes.
- `tools/quality.ps1 -LowMemory` passes, including formatting, tests, build, static analysis, vulnerability scan, tracked-capture checks, and whitespace validation.
- `git diff --check` passes.

## Live validation status

No live or public-server validation was performed. Physical-flow overlap timing is covered only by deterministic local tests.

## Known limitations and follow-up

Flow identity matching uses the source and destination address pair exposed at the observed packet boundary. It does not add lower-level UDP datagram or RakNet connection identifiers. Real-client transfer reconnect timing remains part of the pending local validation work.
