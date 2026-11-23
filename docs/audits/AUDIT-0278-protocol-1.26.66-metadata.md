# AUDIT-0278 - Update Bedrock 1.26.66 metadata baseline

- Commit subject: `Update Bedrock 1.26.66 metadata baseline`
- Change type: `protocol metadata | tests | documentation`
- Runtime impact: `client metadata only`

## Intent

Resolve protocol-update issue [#12](https://github.com/NhanAZ/BedrockDebugProxy/issues/12) after reviewing Mojang's stable `v1.26.66` schema release.

## Evidence and scope

Mojang's `v1.26.66` release declares network protocol `2225`. Its immutable release commit and the reviewed `v1.26.62` commit resolve to the same complete repository tree, so this update changes the game-version metadata without changing packet definitions, protocol ID, or capture schema. The full evidence and license review are in [`docs/research/protocol-1.26.66.md`](../research/protocol-1.26.66.md).

Set the in-tree gophertunnel `CurrentVersion` to `1.26.66`, keep `CurrentProtocol` at `2225`, and retain the reviewed 1.26.60 packet codec. Update current-baseline documentation and metadata tests. No Mojang schema source is copied.

The paired consumer change is BedrockPackArchive commit `04f2e7067a2d07bb759ff5171071099cb7d64451`. It adds `1.26.66` stamped-development metadata coverage and updates compatibility documentation. Its importer production code and archive contract are unchanged.

## Validation

- `go test -run 'TestProtocolVersion12666|TestMojangProtocolWatcherReleaseMetadata' -count=1 ./cmd/bedrock-debug-proxy ./internal/bedrock` passed.
- `tools/quality.ps1 -LowMemory` passed, including tests, build, static analysis, vulnerability scan, license inventory, and whitespace checks.
- BedrockPackArchive `go test ./...`, `go vet ./...`, and `go build ./...` passed. The exact importer fixture against the stamped producer binary is a required post-commit check and will be recorded in the paired validation follow-up.
- No live Bedrock client or server was used.

## Limitations

This is a metadata update based on the identical complete official schema trees and shared protocol ID. It does not establish retail-client or live-server interoperability. The exact producer-to-consumer synthetic importer check remains pending until a clean stamped binary can be built from this commit.
