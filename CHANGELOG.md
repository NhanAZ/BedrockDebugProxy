# Changelog

All notable changes to BedrockDebugProxy are documented in this file. Entries are grouped by user-visible or release-relevant change rather than repeated for every internal commit. Detailed evidence and rationale remain in commit audit records under [`docs/audits/`](docs/audits/).

## [Unreleased]

### Documentation and maintenance

- Added weekly Dependabot checks for the root Go module and GitHub Actions, with grouped non-major Go updates and no automatic merge.
- Added a weekly Mojang protocol release watcher that opens a deduplicated investigation issue without changing protocol code automatically.

## [0.2.0] - 2026-09-16

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

- Initial public release.
