# AUDIT-0213 - Verify stable maintenance on local and public servers

- Commit subject: `docs: record stable maintenance server validation`
- Change type: documentation and paired validation record
- Runtime impact: none

## Tested revisions and preservation

The stamped runtime is `414eeb97157416eed8dfdc85f872101f40dd74af`, version `dev.12652.validation`, built with Go 1.27.2. It contains the separately reviewed security maintenance `a11ef5324b3ab37ff0f1b5679f75e766fcf30b95`, stable metadata `27325cecaf36e9bf90a684742fc31c1973f611b6`, and generated-skin correction. Client game metadata is 1.26.52 and protocol remains 2193.

The paired private BedrockPackArchive consumer is `8afee20944f45fad3faa545e7cdb13d923a26108`. Its preceding security graph commit is `816d23aba51f95c5799a256389cf22c5526d9087`. Both repositories have separate local commits. Source commits have not been pushed and post-push CI is not run.

The owner's six original uncommitted hot-fix files were copied outside the checkout, hash-checked, and saved with their binary diff and base revision. Recoverable stash `5d2b3ba84af769aceb988d1c67fdf682451152b9` remains available. The fixes were discarded from the active checkout before research. The skin correction was then independently reproduced and tested. The owner-controlled PocketMine source checkout retains its original unrelated changes.

## Automated checks

- The canonical formatter and `tools/quality.ps1 -LowMemory -Race` pass on the tested runtime. This includes root tests with the race detector, local transfer and forwarding regressions, viewer asset and HTTP tests, JavaScript syntax, module checks, build, license inventory, static analysis, vulnerability scan, tracked-artifact policy, and whitespace checks.
- Static analysis reports zero issues. Govulncheck reports zero reachable vulnerabilities and one vulnerability in imported or required code that is not called.
- Replacement gophertunnel minecraft, NBT, and packet tests pass with the race detector. Replacement go-raknet's complete race-enabled suite passes.
- BedrockPackArchive tests, vet, build, and diff checks pass. Its eight synthetic combinations cover old and current producer metadata, two language fixtures, and initial and transferred hops. The actual stamped runtime verifier accepts every case. Archive hashes, root and nested manifests, plaintext contents.json, declared synthetic key counts, source event sequences, and index provenance remain preserved. No importer production change is needed.
- The watcher regression failed before separating client metadata from the schema baseline. All seven mocked cases and a read-only real release check pass after correction. No public issue was created.

## Live synthetic sessions

Each session uses the same stamped runtime. A local self-signed protocol client connects to an explicitly unauthenticated loopback listener. Public upstream connections use the owner's cached Microsoft authentication. The client accepts packs, completes normal spawn, reads post-spawn packets for three seconds, and closes. The selected default account was restored after exercising the second cached account on Enchanted.

The isolated Axolotl-PM server uses the owner's existing source checkout at `323aa03de97ec50d8580e03301bf3ef7b5276cb6`, with its pre-existing working-tree changes retained. Its runtime is 5.49.2+dev, server game version 26.50, protocol 2193, and BedrockProtocol `0e13c077d1060ae9e293908eff81fbef600ead5c`. Disposable data and an empty plugin directory are separate from the owner's server data. The test server was stopped after validation.

| Target | Private unchanged capture | Events | Pack archives | Observed flow |
| --- | --- | ---: | ---: | --- |
| Local Axolotl-PM | [session-20261014T180004Z](https://github.com/BedrockDebugProxy-Captures/session-20261014T180004Z/releases/tag/capture-archive) | 481 | 0 | Generated skin login, empty pack negotiation, spawn, both directions, closure |
| The Hive | [session-20261014T180032Z](https://github.com/BedrockDebugProxy-Captures/session-20261014T180032Z/releases/tag/capture-archive) | 16,271 | 25 | Authenticated login, packs, spawn, both directions, closure |
| Galaxite | [session-20261014T180204Z](https://github.com/BedrockDebugProxy-Captures/session-20261014T180204Z/releases/tag/capture-archive) | 72,490 | 20 | Authenticated login, packs, spawn, both directions, closure |
| Lifeboat | [session-20261014T180434Z](https://github.com/BedrockDebugProxy-Captures/session-20261014T180434Z/releases/tag/capture-archive) | 4,559 | 11 | Authenticated login, packs, spawn, both directions, closure |
| Mineville Zeqa | [session-20261014T180546Z](https://github.com/BedrockDebugProxy-Captures/session-20261014T180546Z/releases/tag/capture-archive) | 33,127 | 9 | Authenticated login, URL-backed packs, spawn, both directions, closure |
| Enchanted | [session-20261014T180839Z](https://github.com/BedrockDebugProxy-Captures/session-20261014T180839Z/releases/tag/capture-archive) | 202 | 1 | Authenticated hub login, URL-backed pack, spawn, both directions, closure |

All six captures are closed and pass integrity verification and every automatic report requirement. The total is 127,130 events and 66 pack archives. They contain zero decode errors, unknown packet events, blocking errors, capture write errors, dropped events, and truncated events. Expected read-side termination evidence remains visible. Automatic artifact folders complete at the final sequence for every capture. Plaintext resource-pack derivation was disabled.

The six sanitized reports retain `tested_revision` equal to the runtime above. Every report is `incomplete` with `normal_session=not_observed`, because a synthetic client does not provide retail gameplay evidence. The explicit synthetic spawn and bidirectional checks pass. Pack exchange passes for all five public targets. Local pack exchange is marked not observed because no archive was offered. Enchanted transfer following is marked not observed.

## Archive verification

Each session repository is private. The complete unchanged directory, including automatic artifacts, is packaged as its own ZIP at the capture-archive release. The README records the directory session ID, manifest UUID, filename, and SHA-256. Each uploaded ZIP was downloaded and hash-checked before the capture citation was added. Sanitized reports are attached to the matching private releases. All six report downloads match their prepared hashes, runtime revision, incomplete status, and passing automatic requirements. Original local captures remain retained.

| Session suffix | ZIP bytes | Verified SHA-256 |
| --- | ---: | --- |
| 20261014T180004Z | 968,146 | `9b8b0bf58950161e93f0c3b7f07033f028e3592a03a7c15412d15fcb938237d5` |
| 20261014T180032Z | 128,865,727 | `ce8b789aa398243ff81fb3d109d6355cedc3518934a528d3593ebbe623b2a675` |
| 20261014T180204Z | 290,069,101 | `521611b0c4c3c1233fe29c310c7f8cba67bf31c18845a4601a7864381e214888` |
| 20261014T180434Z | 35,478,502 | `26b70d432e0a260aa5a881ffe8d845af0683a37488161f6d6ab9d4b4ccc5e975` |
| 20261014T180546Z | 335,879,553 | `702ca151ddc38f28ec64d79351a5f747fa0cc5c9b35e8fb23a2a47e33ac32f66` |
| 20261014T180839Z | 5,041,615 | `dddc049087517b6330533dbf5496551f8bc67fcc0662458188ef8ac4c08dbe2b` |

## Limits and human validation

No retail client gameplay, long-running stability session, live NetherNet route, Enchanted minigame transfer, or resource-pack decryption is observed. CubeCraft is not connected because it is a conditional release target with a separate maintainer risk gate. No official release is requested or created. The six-server retail release matrix remains incomplete.

For human validation, join the owner-controlled Axolotl server using the retail 26.52 client through this runtime. Check login, pack negotiation, visible spawn, and movement while retaining the closed capture and server log if a stage fails. Release validation additionally needs the documented six-server retail matrix and Enchanted minigame transfer.

This commit records Markdown evidence only. It preserves the runtime facts and report revisions. The current release is 0.3.0. The accumulated compatibility and security fixes justify a future release checkpoint after retail validation, without turning this update task into release publication.
