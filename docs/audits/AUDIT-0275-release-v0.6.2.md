# AUDIT-0275 - Validate the BedrockDebugProxy v0.6.2 release

- Commit subject: `Prepare BedrockDebugProxy v0.6.2 release`
- Change type: `release documentation`
- Runtime impact: `no`

## Intent

Record the validation evidence for BedrockDebugProxy v0.6.2 and its paired BedrockPackArchive v0.2.1 release.

## Scope and provenance

The validated runtime revision is `39b29f141b918ad0b00cab2cb90fbb862aff56af`. It includes the per-hop connection deadline from [`1140a2f`](https://github.com/NhanAZ/BedrockDebugProxy/commit/1140a2f2a9f33866d4f4b31496bf6dfc1c3b7bb9), the bounded HTTPS NetherNet probe from [`eac0f85`](https://github.com/NhanAZ/BedrockDebugProxy/commit/eac0f855d33721b79329160b35132ee156c03ce2), and the Linux timeout-test fixes in [`a361ba2`](https://github.com/NhanAZ/BedrockDebugProxy/commit/a361ba2a6e95d95e025ecd0176d24c87c57c526d) and `39b29f141b918ad0b00cab2cb90fbb862aff56af`.

The paired BedrockPackArchive consumer commits are [`56d43b7`](https://github.com/NhanAZ-Tools/BedrockPackArchive/commit/56d43b706608d0a9d04bd8b1129639fda263e34e) and [`e0c08f4`](https://github.com/NhanAZ-Tools/BedrockPackArchive/commit/e0c08f4a2f22cc40c2382377bf8f3b62d13bb48f). Its synthetic importer compatibility check passed against the stamped BedrockDebugProxy v0.6.2 binary. Capture schema, archive identity, and importer contract remain unchanged.

The acquisition catalog records `Rob a Mob` and `Treasure Hunt` as Enchanted child destinations with independent Experience selectors and pack targets. This preserves the parent relationship without assuming a fixed form action or transfer.

## Validation

- Automated checks: `tools/quality.ps1 -Race -LowMemory` passed on the runtime revision. The [hosted GitHub Actions run](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/44881075259) passed on Linux and Windows.
- Build: `tools/build.ps1 -Version 0.6.2` produced a binary stamped with BedrockDebugProxy v0.6.2, runtime revision `39b29f141b918ad0b00cab2cb90fbb862aff56af`, Bedrock 1.26.62, and protocol 2225.
- Six-server matrix: `tools/check-release-readiness.ps1` passed all six reports for the exact runtime revision. Every session used authenticated Go Golem, completed offered packs, reached lobby spawn, retained bidirectional decoded traffic for 30 seconds, and closed with verified artifacts. Encrypted offers were fully decrypted.

| Server | Offered packs | Encrypted packs decrypted |
| --- | ---: | ---: |
| The Hive | 23 | 23 |
| Galaxite | 21 | 20 |
| Lifeboat | 13 | 13 |
| Mineville Zeqa | 11 | 0 |
| Enchanted | 1 | 0 |
| CubeCraft | 5 | 5 |

- Focused Enchanted route: the stamped proxy followed the current Dragons form choice through HTTPS NetherNet, received all 10 child archives, and reached child spawn. The closed capture and private archive are recorded in [AUDIT-0274](AUDIT-0274-transfer-nethernet-probe-budget.md).
- Consumer compatibility: `go test -count=1 ./...`, `go vet ./...`, `go build ./...`, and the exact producer fixture passed in BedrockPackArchive. Its current compatibility baseline is BedrockDebugProxy v0.6.2.

## Limitations

The Go Golem sessions establish the captured protocol flows. They do not establish retail rendering or gameplay. CubeCraft remained at its lobby and no minigames were entered. The Enchanted child-route check covered the observed Dragons transfer only.
