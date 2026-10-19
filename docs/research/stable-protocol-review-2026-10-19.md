# Stable protocol review on 2026-10-19

The maintainer requested a protocol update and selected the latest stable release. The source checkout already selects Bedrock 1.26.52 and network protocol 2193. Current primary sources support retaining those values.

## Evidence and decision

Sources below were read on 2026-10-19. GitHub release and tag APIs were also checked directly.

| Claim | Source and revision | Decision |
| --- | --- | --- |
| The latest non-prerelease Mojang schema is v1.26.51, protocol 2193 | [Stable release](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.51), published on 2026-09-16 and resolved to commit [8667debe0adad27de6c692c88942c489a1193206](https://github.com/Mojang/bedrock-protocol-docs/tree/8667debe0adad27de6c692c88942c489a1193206) | Retain CurrentProtocol 2193. |
| Stable 1.26.51 has no packet schema changes from 1.26.50 | The official [protocol changelog](https://mojang.github.io/bedrock-protocol-docs/changelog/) lists zero changes for stable 1.26.51 | Retain the reviewed 1.26.50 codec composite and local observation hooks. |
| The current published game hotfix is 26.52 | The official [release index](https://feedback.minecraft.net/hc/en-us/categories/115000410252-Knowledge-Base) lists the [26.52 hotfix](https://feedback.minecraft.net/hc/en-us/articles/49175370527501-Minecraft-Bedrock-Edition-26-52-Hotfix-Changelog), published on 2026-09-25. The installed Windows package is Microsoft.MinecraftUWP 1.26.5203.0 | Retain CurrentVersion 1.26.52. A game hotfix does not supply an exhaustive packet schema. |
| Newer schemas are Preview releases | The [Mojang release list](https://github.com/Mojang/bedrock-protocol-docs/releases) marks v1.26.70-preview.20, protocol 2235, and the 1.26.60 previews as prereleases | Keep Preview layouts outside the selected stable codec. |
| The newest gophertunnel tag remains v1.62.0 | Tag API and [7a556a07335b663744b50d38062636ad8283f314](https://github.com/Sandertv/gophertunnel/tree/7a556a07335b663744b50d38062636ad8283f314) | Its reviewed protocol corrections are already recorded in the in-tree patch provenance. No dependency replacement is needed. |
| Upstream master still selects protocol 2193 and game version 1.26.50 | [minecraft/protocol/info.go at 80c811b6186016b3860c358368cfa47e507f26e9](https://github.com/Sandertv/gophertunnel/blob/80c811b6186016b3860c358368cfa47e507f26e9/minecraft/protocol/info.go) | This executable implementation independently agrees on the protocol ID. Preserve the separately reviewed local 1.26.52 hotfix metadata. |

The protocol watcher was run without CreateIssue. It reported that no stable release newer than the project baseline was detected. No issue was created.

## Scope and validation

This review changes documentation only. Packet IDs, serializers, login, transport, encryption, compression, resource-pack handling, capture fields, CLI behavior, and the viewer remain unchanged. No new wire definition or upstream source is imported. Mojang material is specification evidence. The existing gophertunnel MIT provenance remains applicable.

The focused replacement-module protocol and packet tests and the canonical tools/quality.ps1 -LowMemory gate passed. The gate covered formatting, script and viewer syntax, module and license checks, root tests, build, static analysis, vulnerability analysis, and whitespace. Static analysis reported zero issues and the vulnerability scan found no called vulnerabilities. The existing root protocol fixtures check protocol 2193, 1.26.52 metadata, packet registration, exact packet bodies, inventory hand encoding, sub-chunk height maps, and the observed cache-hash count.

The sibling BedrockPackArchive checkout was inspected and is clean. It uses this replacement module and already matches the selected stable protocol. This documentation-only review changes no producer or importer contract and requires no consumer patch.

There is still no released 1.26.52 Mojang schema that proves an exhaustive absence of wire differences. This source review does not establish new public-server or retail-client compatibility. Keep any subsequent CLI sessions scoped as synthetic-client evidence and preserve the existing official release gate.
