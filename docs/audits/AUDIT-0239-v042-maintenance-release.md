# AUDIT-0239 - Update the Go security baseline before release

The resumed v0.4.2 release gate found [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809), published October 21, after the October 20 live matrix. The advisory identifies compromised intermediate-certificate trust in crypto/x509 before Go 1.27.3. The scanner found Certificate.Verify reachable. This change updates the preferred toolchain from Go 1.27.2 to 1.27.3 and its documented baseline. No packet definition, dependency module version, capture schema, or product interface changes.

The old runtime is 244841160279ebbda0af02369737940f01be4254. Its six canonical reports passed the release checker and retain that revision. All six private capture archives and their separate driver source assets were verified against local SHA-256 digests. Originals remain unchanged locally.

| Server | Private evidence | Packs | CLI stability |
| --- | --- | --- | --- |
| The Hive | [session-20261020T155538Z](https://github.com/BedrockDebugProxy-Captures/session-20261020T155538Z/releases/tag/capture-archive) | 25 | 35 seconds |
| Galaxite | [session-20261020T155810Z](https://github.com/BedrockDebugProxy-Captures/session-20261020T155810Z/releases/tag/capture-archive) | 20 | 35 seconds |
| Lifeboat | [session-20261020T160143Z](https://github.com/BedrockDebugProxy-Captures/session-20261020T160143Z/releases/tag/capture-archive) | 11 | 35 seconds |
| Mineville Zeqa | [session-20261020T160537Z](https://github.com/BedrockDebugProxy-Captures/session-20261020T160537Z/releases/tag/capture-archive) | 9 | 35 seconds |
| Enchanted | [session-20261020T161718Z](https://github.com/BedrockDebugProxy-Captures/session-20261020T161718Z/releases/tag/capture-archive) | 1 hub and 7 child | 35 seconds on child |
| CubeCraft | [session-20261020T161951Z](https://github.com/BedrockDebugProxy-Captures/session-20261020T161951Z/releases/tag/capture-archive) | 5 | 35 seconds |

Enchanted authenticated both hops, selected the server's OP SMP form choice, followed Transfer, completed child packs, and spawned. Its late generic upstream connection termination occurred about 24 seconds after the CLI completed and closed. The raw error remains preserved. The optional route ping was marked best-effort and followed by successful child login. CubeCraft used the normal observation path after explicit maintainer risk acceptance. This is not server permission or a policy exception.

The initial Mineville catalog request ended with EOF before a capture or server connection existed. A normal retry succeeded. The earlier [Enchanted diagnostic](https://github.com/BedrockDebugProxy-Captures/session-20261020T160959Z/releases/tag/capture-archive) remains incomplete. The temporary CLI driver missed a decorated OP SMP label. Its focused selector correction passed decorated-label, ambiguity, lookalike, destination, and loopback tests. Original v1 and corrected v2 sources, module files, build information, and binary hashes remain attached to the matching private archives. No product workaround was added.

Earlier runtime CI passed in [run 39452469099](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/39452469099). Browser checks at 34472e2 covered live closure, filters, details, and ordinary, empty, and absent blobs. A closed-capture check at 2448411 confirmed the empty-blob display. Console errors were absent. BedrockPackArchive commit 2e5c9e2fcbf21812e477198760ffcf994144dbb6 passed its paired synthetic compatibility checks, race tests, vet, and build against the original stamped runtime.

The new toolchain requires fresh exact-revision CI, a stamped build, and a six-server matrix. The maintainer accepted the CubeCraft risk for this additional session. Historical reports must not be relabeled. CLI observations establish neither retail rendering nor gameplay. Live decryption was disabled. Publication stays blocked until the security-updated candidate passes every release gate.

Go 1.27.3 passed tools/quality.ps1 -Race, including formatting, module integrity, race tests, build, zero static-analysis findings, and zero called-function vulnerabilities. The advisory remains visible in the historical failed Go 1.27.2 scan. No scanner exclusion was added.
