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

Before editing code, trace the relevant data path and inspect call sites, shared helpers, interfaces, similar implementations, tests, documentation, and useful Git history. Search for existing functionality before creating another implementation. Reuse or improve the established path when it fits rather than introducing duplicate or competing logic.

When uncertainty remains between adding code and investigating further, investigate further. When both a small and a large patch solve the verified root cause, prefer the small patch.

## Issue and pull request triage

Treat an issue as a report to investigate, not an implementation order. Determine whether it is reproducible, intentional, duplicate, in scope, supported by evidence, and compatible with the project mission. It is valid to request more information or recommend closing an issue when the technical reason is explicit.

Review pull requests against the current implementation and architecture before judging their diff. Check correctness, duplicate functionality, unrelated changes, dependencies, protocol assumptions, capture fidelity, cross-server compatibility, tests, and maintenance cost. Compilation alone is not sufficient evidence for approval. Request focused, actionable changes or recommend closing a pull request when warranted.

## Protocol research

Use primary sources whenever possible. Inspect the actual source revision, protocol definitions, tests, and wire behavior. Distinguish confirmed behavior, reasoned inference, and unknown behavior in code comments and documentation.

Cross-check protocol work against multiple independent implementations or captures when practical. Useful references include gophertunnel, go-raknet, Cloudburst Protocol and ProxyPass, Kas-tle ProxyPass, PrismarineJS bedrock-protocol, Endstone protocol-dumper and spyglass, bedrocktool, Dragonfly, and observed Bedrock Dedicated Server behavior. If sources disagree, investigate protocol versions, optional fields, experiments, feature flags, transport differences, and implementation workarounds before choosing a definition.

Never infer packet IDs, field types, serialization order, version gates, or required packet order merely because a proposed layout appears plausible. Every new protocol behavior needs a traceable source or capture. Record why the selected interpretation fits the evidence.

Record important protocol findings in `docs/research/`. Include source URLs, revision identifiers, access dates when useful, relevant license information, and the impact on this project. Do not let important findings exist only in a chat, commit message, or code comment.

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

Redaction must be an explicit export operation. Preserve the original local capture unless the operator requests deletion. Documentation and test fixtures must use synthetic or deliberately public data.

Do not weaken authentication, encryption, or certificate validation merely to simplify implementation. Document unavoidable trust boundaries and local interception requirements.

## Dependencies and source reuse

BedrockDebugProxy is licensed `GPL-3.0-or-later`. The repository is private only during development and is intended to become public when ready. Do not describe private status as an all-rights-reserved license or as an exception to the GPL.

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

Fuzz parsers and decoders that consume untrusted network or capture data when practical. Bound memory, disk, and goroutine growth without hiding the fact that a limit was reached.

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

## Change control

Do not weaken capture fidelity, raw-data retention, security safeguards, schema compatibility, license rules, or validation requirements for implementation convenience.

Material architecture decisions belong in `docs/decisions/`. Significant changes to this guide must be deliberate, documented, and committed separately when practical.

Packet mutation, dropping, injection, replay, cheat behavior, and exploit tooling are outside the initial scope. Add them only after a documented debugging need, security review, and explicit project decision.
