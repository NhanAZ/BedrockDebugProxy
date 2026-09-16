# BedrockDebugProxy Agent Guide

## Mission

BedrockDebugProxy is a high-fidelity observation and research tool for Minecraft Bedrock traffic. Its primary users are AI agents, followed by the project owner and protocol engineers.

The central principle is "Capture once, analyze many times." Prefer preserving evidence over producing short or attractive logs.

## Decision priorities

Make technical decisions in this order.

1. Preserve observable data and provenance.
2. Produce deterministic, structured, machine-readable captures.
3. Keep packet ordering, direction, timing, and connection context intact.
4. Expose decode failures and protocol uncertainty instead of hiding them.
5. Keep protocol and transport components replaceable.
6. Favor maintainability and documented behavior over clever shortcuts.

Do not select an architecture only because an existing proxy uses it. Research current implementations, compare trade-offs, and record consequential choices before committing to them.

## Maintainer posture

Maintain the project conservatively. Prefer the smallest focused change that fixes the demonstrated root cause. Do not turn a bug fix, protocol update, issue, or pull request into a broad refactor unless the existing design makes the focused change unsafe.

Do not add a feature, public API, abstraction, helper, framework, compatibility layer, subsystem, or dependency because it may be useful later. Require a current use case and evidence that existing project mechanisms cannot address it cleanly. Feature additions and removals are material decisions and must not be hidden inside unrelated fixes.

Keep the product CLI small. Do not add specialized download, extraction, logging, or one-purpose commands merely because a reference project has them. Prefer retaining observable evidence in the canonical session so existing inspect, analyze, explain, and export workflows can derive later views from one capture.

Before editing code, search [`docs/fix-history.md`](docs/fix-history.md) for the reported symptom and inspect the linked decisions, research notes, captures, tests, and Git history. Trace the relevant data path and inspect call sites, shared helpers, interfaces, similar implementations, and documentation. Classify the issue as a known fix, regression, limitation, duplicate, or new root cause before changing code. Reuse or improve the established path when it fits rather than introducing duplicate or competing logic.

Treat every requested solution, including one from the project owner, as a proposal to evaluate rather than proof that the requested implementation is correct. Confirm the underlying need, project fit, existing behavior, evidence, risk, and smallest maintainable solution before changing code. Ask a focused question when missing information would materially change the result. Otherwise proceed with the best-supported narrow solution and state any important assumption. Push back, propose a different approach, reduce scope, or decline implementation when the request conflicts with evidence, project boundaries, security, capture fidelity, licensing, or maintenance goals. Explain the technical reason and give an actionable alternative when possible.

When uncertainty remains between adding code and investigating further, investigate further. When both a small and a large patch solve the verified root cause, prefer the small patch.

## Issue and pull request triage

Treat an issue as a report to investigate, not an implementation order. Determine whether it is reproducible, intentional, duplicate, in scope, supported by evidence, and compatible with the project mission. It is valid to request more information or recommend closing an issue when the technical reason is explicit.

Review pull requests against the current implementation and architecture before judging their diff. Check correctness, duplicate functionality, unrelated changes, dependencies, protocol assumptions, capture fidelity, cross-server compatibility, tests, and maintenance cost. Compilation alone is not sufficient evidence for approval. Request focused, actionable changes or recommend closing a pull request when warranted.

## Protocol research

Use primary sources whenever possible. Inspect the actual source revision, protocol definitions, tests, and wire behavior. Distinguish confirmed behavior, reasoned inference, and unknown behavior in code comments and documentation.

Cross-check protocol work against multiple independent implementations or captures when practical. Begin with the matching released schema in Mojang `bedrock-protocol-docs` when one exists, then compare gophertunnel, go-raknet, Cloudburst Protocol and ProxyPass, Kas-tle ProxyPass, PrismarineJS bedrock-protocol and minecraft-data, Endstone spyglass, endstone, endweave, and protocol-dumper, axolotl-pm PocketMine-MP and BedrockProtocol, Altay, BetterAltay, `bedrock-tool/bedrocktool`, Dragonfly, and observed Bedrock Dedicated Server behavior. Use the source roles and review requirements in [`docs/protocol-updates.md`](docs/protocol-updates.md). If sources disagree, investigate protocol versions, optional fields, experiments, feature flags, transport differences, generated-code lag, fork lineage, and implementation workarounds before choosing a definition.

Never infer packet IDs, field types, serialization order, version gates, or required packet order merely because a proposed layout appears plausible. Every new protocol behavior needs a traceable source or capture. Record why the selected interpretation fits the evidence.

Record important protocol findings in `docs/research/`. Include source URLs, revision identifiers, access dates when useful, relevant license information, and the impact on this project. Do not let important findings exist only in a chat, commit message, or code comment.

Before adding or relying on an external citation, verify that the exact URL exists and that its current content supports the stated claim. Prefer official sources and immutable commit or tag permalinks for revision-specific behavior. Record the repository, file path, revision, and evidence summary when the claim is important. A successful HTTP response alone is insufficient. If the original evidence cannot be recovered, mark the claim unverified instead of substituting a nearby URL or inference.

Protocol-specific behavior must not leak into generic capture storage without a documented reason. Keep protocol versions explicit. Unknown packet IDs and partially decoded payloads are valid evidence and must remain representable.

## Capture integrity

Never silently discard raw data, malformed input, unknown packets, decode errors, compression details, encryption state, batching information, or timing information when the layer makes them observable.

Decoded values supplement raw evidence. They do not replace it. Store raw bytes once using content-addressed blobs when practical and reference them from ordered events.

Capture timestamps must be monotonic within a process and also include a wall-clock representation suitable for cross-system analysis. Preserve the original event order even when timestamps collide.

Every lossy operation, redaction, truncation, sampling rule, or size limit must be explicit in capture metadata and visible to the operator.

Capture schema changes require a version bump, compatibility notes, fixtures, and tests for both writing and reading. Do not rewrite existing captures in place.

## Security and privacy

Treat authentication tokens, private keys, chain data, server addresses, device identifiers, chat, and captured payloads as sensitive.

Never commit live credentials, user captures, decrypted private resource packs, or generated authentication caches. Default capture directories and secrets must be ignored by Git.

Resource-pack decryption is a conservative project boundary. Production code may use only the archive and `ContentKey` delivered by the upstream server on the current accepted connection. Do not add an offline decrypt command, arbitrary remote pack retrieval, a shared key database, key guessing, brute force, recovery, authentication bypass, automated upload, or redistribution without explicit maintainer review of the technical need, legal risk, and documentation. Preserve the encrypted archive before deriving plaintext. This invariant records actual scope and is not a claim that every use is legally authorized.

Redaction must be an explicit export operation. Preserve the original local capture unless the operator requests deletion. Documentation and test fixtures must use synthetic or deliberately public data.

Do not weaken authentication, encryption, or certificate validation merely to simplify implementation. Document unavoidable trust boundaries and local interception requirements.

## Dependencies and source reuse

BedrockDebugProxy is licensed `GPL-3.0-or-later`. The repository is private only during development and is intended to become public when ready. Do not describe private status as an all-rights-reserved license or as an exception to the GPL.

Captured and decrypted third-party content is not automatically covered by the project GPL. Keep the source-code license, dependency licenses, capture ownership, and artifact redistribution rights distinct in code, documentation, release material, and support responses.

Check dependency licenses, maintenance status, supported protocol versions, and extension points before adoption. Prefer stable public APIs, but retain access to raw boundaries required for debugging.

Ideas may be reimplemented after study. Source code may be copied only when its license permits the intended use and attribution requirements are satisfied. Record copied or substantially adapted code in `THIRD_PARTY_NOTICES.md` with its source revision and license.

Keep independent implementation, dependency use, research influence, and copied or substantially adapted source distinguishable. Update provenance in the same commit that introduces reused code. Preserve required copyright, license, and notice text in source and binary distributions. A compatible dependency does not make unlicensed source reusable, and code whose license would add obligations beyond `GPL-3.0-or-later` requires an explicit project decision before reuse.

Do not copy code from a source whose license is absent, unclear, or incompatible. Private repositories owned by the project owner remain separate works unless their code and license are deliberately imported and documented.

## Testing and validation

Every behavior change needs validation proportional to its risk. Prefer small deterministic unit tests, binary fixtures, round-trip tests, malformed-input tests, and local loopback integration tests.

Transport, compression, encryption, batching, framing, packet decoding, resource-pack reconstruction, and schema compatibility require focused regression coverage when changed.

Run formatting, unit tests, static analysis, build checks, and `git diff --check` before committing. A successful local build is not evidence of compatibility with a real Bedrock client or public server. Keep manual and live verification status explicit.

Use `tools/quality.ps1` as the project-wide pre-commit quality gate and `tools/format.ps1` as the canonical formatting command. Do not replace, bypass, or duplicate this workflow in an issue-specific script. Add a linter only when it provides actionable signal for this codebase, and prefer resolving valid findings over broad exclusions. Keep tool versions pinned and review version changes separately from unrelated behavior changes.

After a commit changes protocol handling, networking, resource packs, or observable behavior, provide a short human validation plan. Name the server or server category, the flow to exercise, the packets or behavior to observe, the expected result, and the capture or log needed if it fails. Automated checks and real-world validation must be reported separately.

The Hive is the minimum live baseline for every runtime-affecting development change and pull request. Require a machine-readable report generated from a stamped binary and closed capture for the exact revision. A specific flow must be exercised directly. A report that says `not_observed` remains incomplete. Do not approve or merge a runtime-affecting change while required validation is pending.

Every official release requires revision-matched reports for The Hive, Galaxite, Lifeboat, Mineville Zeqa, and Enchanted. Record external outages or unavailable tests explicitly. A failed server starts root-cause investigation and does not authorize a server-specific workaround. Follow `docs/validation.md` and `docs/server-targets.md` for the report workflow, endpoint references, and sensitive-data boundary.

CubeCraft is excluded from active validation, release gates, pull request recommendations, and proposed test targets because its published policies prohibit Bedrock proxy use. Existing CubeCraft research is historical evidence only. If the maintainer asks to test CubeCraft, including `play.cubecraft.net:19132` or `experience:CubeCraft`, warn about the published policy and possible rejection, kick, or ban before proceeding. This is an internal agent workflow rule and must not become a product runtime warning.

Fuzz parsers and decoders that consume untrusted network or capture data when practical. Bound memory, disk, and goroutine growth without hiding the fact that a limit was reached.

## Release stewardship

When the maintainer says they want to release, treat that as a request to perform a complete release-readiness audit, not as a request to immediately create a tag. The agent owns every automatable step: inspect the diff and commits since the latest release, propose or confirm the version, check scope and provenance, run quality and build gates, verify CI, create the stamped candidate binary, inspect completed captures, generate sanitized reports, run the five-server gate, prepare checksums and release notes, and verify the published release.

Ask the maintainer to perform only actions that require a real Minecraft client, account interaction, judgment of visible gameplay, unavailable signing authority, or another human-only boundary. Give one short concrete live-test instruction at a time when practical. After the maintainer reports completion, inspect the generated capture and machine-readable evidence rather than asking them to run validation scripts manually. If any gate is incomplete, explain the exact blocker and do not publish. If the maintainer explicitly requested the release, the version is settled, every required gate passes, and no unresolved review issue remains, proceed with the tag and GitHub release workflow in `docs/releasing.md` without asking them to repeat already established release intent. After the published release is verified, synchronize all Git-tracked branches and tags with the private `NhanAZ/BedrockDebugProxy-Backup` recovery mirror and verify its refs before reporting the release as complete. Do not place captures, credentials, private keys, or build output in that mirror.

After a substantial capability, protocol update, capture schema or storage change, networking or authentication change, compatibility fix, or coherent series of large commits, compare the current state with the latest release. Consider whether a release checkpoint would create a useful tested and reversible boundary before more work accumulates. Recommend pausing for release validation when justified, but do not stop unrelated work, tag, publish, or turn commit count alone into a release requirement. State why a checkpoint is useful, what remains unverified, and the minimum human validation still required.

## Compatibility and protocol updates

Keep protocol support isolated behind explicit interfaces and version metadata. Before updating protocol definitions, read `docs/protocol-updates.md`, inspect upstream changes, identify local extension points, and establish a passing baseline.

Prefer focused updates that preserve BedrockDebugProxy hooks. Do not copy a large upstream tree blindly. Validate packet IDs, field changes, compression negotiation, resource-pack flow, authentication, and representative capture fixtures.

When a new Bedrock release is not fully understood, capture unknown data faithfully and mark decode confidence instead of guessing fields.

Do not assume all Bedrock servers use the same valid packet sequence. Featured Experiences, Creator Experiences, Bedrock Dedicated Server, and other server software may delay, omit, reorder, or add packets where the protocol permits it. Avoid order-dependent logic unless the protocol requires that order and the requirement has evidence. State machines must tolerate the valid sequences demonstrated by supported server categories.

For a bug that appears after a protocol update, verify packet definitions, field order, optional fields, version gates, decoding boundaries, and server-specific behavior before adding a compatibility workaround. Fix protocol understanding before patching symptoms when the evidence supports it.

## Documentation

Source comments, documentation, CLI output, logs, schemas, and generated project text must be written in English.

Do not use Em dash or En dash characters. Use a Hyphen where punctuation is needed. Use straight double quotes. Avoid unnecessary colons and semicolons in prose, especially label-style fragments.

Documentation must explain observable behavior, trust boundaries, storage formats, protocol assumptions, and validation status. A new contributor or AI agent should be able to understand the data path without reading the entire codebase.

Update documentation in the same commit as the behavior it describes whenever possible.

## Go conventions

Use standard Go style and `gofmt`. Keep packages cohesive and names direct. Accept `context.Context` for cancellable or long-running work. Return errors with useful operation context and preserve underlying errors for inspection.

Avoid package globals for mutable session state. Make ownership and shutdown of files, sockets, channels, and goroutines explicit. Ensure cleanup is idempotent where concurrent failure paths can meet.

Use interfaces at boundaries that genuinely need alternative implementations, such as transport observation, protocol codecs, blob stores, capture sinks, and exporters. Do not create interfaces solely for hypothetical abstraction.

Structured logs must not be the canonical capture. Operator logs may summarize progress, but detailed evidence belongs in the capture model.

## Git workflow

Keep the repository in a reviewable and recoverable state. Commit each coherent stable change after validation. Avoid mixing unrelated changes or creating commits for isolated character edits that belong to one logical change.

Use short imperative commit subjects that describe the change. Do not rewrite shared history or force push without a specific reviewed reason. Before high-risk architecture, protocol, schema, or storage work, commit the current stable state.

Inspect `git status` and the staged diff before every commit. Never discard unrelated user changes. Never commit generated captures, secrets, build outputs, or local authentication state.

After every push, inspect GitHub Actions for the exact pushed commit and wait for every required workflow job to finish. A local quality pass is not a CI result. If a job fails, open its logs, classify the failure, make a focused fix with a new audit record, push again, and repeat until the exact commit is green. Do not declare a pushed change complete while its required workflow is pending or failing.

Every authored non-merge commit must include exactly one commit audit record under [`docs/audits/`](docs/audits/). Create the record before committing and stage it with the coherent change. Use a stable change ID and descriptive filename rather than a commit hash, because the final hash is not known until the commit is created. Git metadata remains authoritative for the hash, author, timestamp, and parent history. The record must state the intent, scope, affected files or interfaces, evidence and provenance, automated validation, live validation status, known limitations, and follow-up work. Documentation-only and formatting-only commits still get a short record that explicitly says they have no runtime impact. Do not rewrite older history to retrofit records; mark the policy start in the first audit record and use the existing Git history, research notes, decisions, and captures for earlier work.

## Change control

Do not weaken capture fidelity, raw-data retention, security safeguards, schema compatibility, license rules, or validation requirements for implementation convenience.

Material architecture decisions belong in `docs/decisions/`. Significant changes to this guide must be deliberate, documented, and committed separately when practical.

Packet mutation, dropping, injection, replay, cheat behavior, and exploit tooling are outside the initial scope. Add them only after a documented debugging need, security review, and explicit project decision.
