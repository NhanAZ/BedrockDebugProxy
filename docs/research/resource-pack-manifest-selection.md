# Resource-pack manifest selection

Accessed 2026-09-25.

## Finding

Minecraft's [Add-Ons manifest reference](https://learn.microsoft.com/en-us/minecraft/creator/reference/content/addonsreference/packmanifest?view=minecraft-bedrock-stable) describes `manifest.json` as the information Minecraft uses to identify a pack. Its `header.uuid` uniquely identifies the pack, and `header.version` identifies its version.

The reported OneBlockOnline failure involved an archive with a nested placeholder `manifest.json` before the pack's root manifest. `packReader.find` matched the first entry by basename, so `resource.Read` parsed the nested identity. The previous temporary capture was removed and cannot be reviewed now. The archive ordering and identity-selection failure are therefore recorded as the reported cause, not as independently reviewable capture evidence.

## Implementation and compatibility

The vendored gophertunnel reader now checks for the exact archive-root path before falling back to the first nested basename match. `readManifest` uses this helper for `manifest.json`; the same helper also makes `level.dat` lookup prefer the archive root. The vendored code is MIT licensed and based on the revisions documented in `THIRD_PARTY_NOTICES.md`.

The BedrockDebugProxy capture schema, resource-pack events, and archive bytes do not change. The companion BedrockPackArchive importer therefore needed no runtime code change. Its compatibility update is commit `95c919474fb2dfa9549527a2c5f6389fdc8f165f` in `NhanAZ-Tools/BedrockPackArchive`. That commit adds a synthetic capture import test which verifies the original root and nested manifest bytes, decrypted archive hash, `contents.json`, per-file key count, and archive-index metadata.

## Validation and limits

`TestResourceReadPrefersRootManifestOverNestedPlaceholder` uses a synthetic ZIP with a nested placeholder before the root manifest and confirms that the root UUID, version, and name are selected. `tools/format.ps1` and `tools/quality.ps1 -LowMemory` passed. BedrockPackArchive `go test -count=1 ./...`, `go vet ./...`, and `go build ./...` passed. Its integration test substitutes a test-only success response for BedrockDebugProxy's external `verify` command, so it covers archive import behavior rather than the verifier itself.

The earlier owner-run OneBlockOnline and The Hive pack-only sessions completed, but their temporary captures were removed and they are not stamped reports for the committed runtime revision. The exact-revision The Hive baseline and a direct OneBlockOnline pack flow remain pending. Synthetic tests do not prove compatibility with a retail client or live server.
