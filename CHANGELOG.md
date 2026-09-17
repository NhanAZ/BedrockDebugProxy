# Changelog

All notable changes to BedrockDebugProxy are documented in this single file. Each release keeps a concise, curated
summary here, while GitHub's compare view and the commit audit records provide the complete history and evidence.
Detailed rationale remains in [`docs/audits/`](docs/audits/).

## [Unreleased]

## [0.2.1] - 2026-09-17

Supported Bedrock version: 1.26.50 (protocol 2193).

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.2.1) | [All changes since v0.2.0](https://github.com/NhanAZ/BedrockDebugProxy/compare/v0.2.0...v0.2.1)

### Added

- Added weekly Dependabot checks for the root Go module and GitHub Actions, with grouped non-major Go updates and no automatic merge.
- Added a weekly Mojang protocol release watcher that opens a deduplicated investigation issue without changing protocol code automatically.

### Changed

- Clarified maintainer review-comment guidance for bot-authored dependency pull requests, including concise natural wording, bot-neutral language, and punctuation.
- Documented the server-policy disclaimer, AI-assisted maintenance, protocol research, and resource-pack encryption review in the project guidance.
- Release notes now link to the relevant section in this changelog and to GitHub's compare view for the complete release commit history.

### Fixed

- Aligned Bedrock 1.26.50 sub-chunk height-map decoding with the released upstream schema, including row-length validation while preserving raw forwarding.

### Documentation and maintenance

- Updated dependency provenance and protocol baselines to record the merged gophertunnel 1.26.50 correction.

## [0.2.0] - 2026-09-16

Supported Bedrock version: 1.26.50 (protocol 2193).

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.2.0) | [All changes since v0.1.0](https://github.com/NhanAZ/BedrockDebugProxy/compare/v0.1.0...v0.2.0)

### Added

- Added Bedrock 1.26.50 support with protocol 2193 metadata and focused packet coverage.

### Changed

- Capture writes now use one bounded, ordered asynchronous writer with blocking backpressure and a never-drop policy, keeping disk work off packet-forwarding callbacks while preserving event order and raw evidence.
- URL-advertised resource packs are prefetched concurrently with bounded workers and committed in the original offer order.
- Pre-spawn packet summaries are shown independently for each transfer hop.

### Fixed

- Packed `PlayerAuthInput` inventory actions now preserve the 1.26.50 `Hand` field during decoding and encoding.
- Coordinated peer-close handling no longer reports expected cancellation and closed-connection races as blocking bridge errors.

### Documentation and maintenance

- Documented the upstream fair-play contribution process for defects found in third-party dependencies.
- Release workflow documentation now requires synchronization with the private source-history backup after a published release.

## [0.1.0] - 2026-09-14

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.1.0)

- Initial public release.
