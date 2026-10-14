# Stable maintenance review on 2026-10-15

The owner requested preservation and removal of the uncommitted hot fixes, a current stable update, and automated server checks using the existing accounts. The original files and binary Git diff were copied outside the checkout and hash-checked before a recoverable Git stash was created. The clean starting revision is `1db39056551b248f80aeedf01eeb1ebc79a4dcbf`.

## Toolchain evidence

The clean baseline passed unit and integration tests, build, viewer JavaScript syntax, license inventory, and static analysis. The canonical quality gate failed at its vulnerability scan using the installed Go 1.27.0. Its JSON evidence identified ten reachable standard-library advisories, including [GO-2026-6617](https://pkg.go.dev/vuln/GO-2026-6617). The official [Go download index](https://go.dev/dl/?mode=json) selects Go 1.27.2 as the latest stable patch. The advisories identify Go 1.27.2 and golang.org/x/net v0.60.0 as fixed versions for the affected HTTP/2 implementation.

Select Go 1.27.2 and x/net v0.60.0, including the required x/crypto v0.57.0, x/sys v0.48.0, and x/text v0.42.0 graph. Review this dependency change separately from Bedrock metadata. Keep the canonical quality workflow and explicit checker pins. Mirror the networking dependency graph in the replacement gophertunnel module and BedrockPackArchive. No observation hook, packet definition, capture schema, or archive contract is changed by this maintenance step.

The first Go 1.27.2 quality run then exposed a tool compatibility failure. golangci-lint v2.13.2 could not read compiler export format 5 and reported a maximum supported format of 4. The latest stable [golangci-lint v2.14.0 release](https://github.com/golangci/golangci-lint/releases/tag/v2.14.0), tag commit `793d05316077345337ae9a11ae05ec8f23edfe1b`, selects x/tools v0.50.0. Its `internal/pkgbits/version.go` explicitly includes V5 and PreserveMethodOrder. Select that released checker pin rather than disabling static analysis or changing the project linters. The linter release and module sources were inspected before changing the pin.

## Stable Bedrock evidence

| Claim | Primary source | Decision and confidence |
| --- | --- | --- |
| The latest game hotfix is 26.52, published on 2026-09-25 | [Official game changelog](https://feedback.minecraft.net/hc/en-us/articles/49175370527501-Minecraft-Bedrock-Edition-26-52-Hotfix-Changelog) and the installed Microsoft.MinecraftUWP package version 1.26.5203.0 | Select 1.26.52 as client game-version metadata. Confirmed game release, not a new packet specification. |
| The newest released stable protocol schema remains v1.26.51, protocol 2193 | [Mojang stable release](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.51), resolved to [8667debe0adad27de6c692c88942c489a1193206](https://github.com/Mojang/bedrock-protocol-docs/tree/8667debe0adad27de6c692c88942c489a1193206), and the release API | Retain protocol 2193 and the reviewed packet codec. There is no official 1.26.52 schema to prove an exhaustive absence of wire changes. |
| Newer 1.26.60 and 1.26.70 schemas are prereleases | [Mojang releases](https://github.com/Mojang/bedrock-protocol-docs/releases), including v1.26.70-preview.20 | Do not substitute Preview layouts for the stable client. |
| The latest stable gophertunnel tag remains v1.62.0 | [7a556a07335b663744b50d38062636ad8283f314](https://github.com/Sandertv/gophertunnel/tree/7a556a07335b663744b50d38062636ad8283f314) and its tag API | Its relevant protocol corrections are already in the documented composite. Retain that reviewed source and its original base declaration instead of overwriting local capture and compatibility hooks. |

The exact source URLs were read on 2026-10-15. Mojang material is specification evidence and is not copied as source. The selected Go dependencies retain their existing compatible licenses.

## Local reproduction

An isolated Axolotl-PM 5.49.2+dev runtime uses the owner's existing source and vendor checkout, protocol 2193, an empty plugin directory, and disposable world data. The baseline packs-only login failed before resource-pack negotiation. The local server explicitly reported an empty SkinArmSizeType value. This reproduces the need behind one discarded skin-default hot fix independently of the old patch. Inspect the matching PHP enum, color parser, and client-data conversion before selecting a correction. Preserve supplied skin fields and cover authenticated and offline default generation.

## Validation boundaries

Run the canonical quality gate and focused replacement-module checks for each runtime step. Use clean stamped binaries for final public-server checks and sanitized reports. Packs-only validates login and the offered pack exchange without world spawn or retail gameplay. Local synthetic clients can additionally exercise spawn and bidirectional forwarding, but are not retail client evidence. CubeCraft is outside this exploratory task because it is a conditional release target with a separate risk acknowledgement. No release is requested.

Live results and paired consumer revisions will be recorded in later audit records after the stamped checks finish. Raw captures remain ignored and local until any capture cited as evidence is privately archived and verified.
