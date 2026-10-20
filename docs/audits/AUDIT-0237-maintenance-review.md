# AUDIT-0237 - Fix maintenance review findings

The maintainer requested a codebase review, focused fixes, removal of unused or repeated code, shorter documentation, and a validated release. The starting tree was clean at v0.4.1, commit 5a58f1358e9f6e3e8371f3231d860dd2a9337127.

## Scope and evidence

The review covered the CLI and release tools, capture writer and readers, analysis and export, artifacts, viewer, authentication, Experience signaling, proxy and transfer paths, Golem integration, and local transport hooks. Existing fix history and call sites were checked before editing. Static analysis found no unused-code findings. Changes address reproduced defects rather than introducing new abstractions or dependencies.

- Capture verification now rejects manifest paths that disagree with the fixed capture v1 layout. A synthetic alternate event file previously passed verification and could replace the real stream in an export. The schema and valid producer output are unchanged.
- The live viewer indexes the final events before external verification stops tailing. Its regression previously returned zero indexed events and blobs despite a verified manifest containing both. Index failures remain visible. The CLI's ineffective refresh after completion is removed.
- The viewer distinguishes an empty blob from an event with no blob reference. JavaScript checks cover omitted and empty previews, nonempty and truncated previews, and absent blobs.
- Account rename refuses existing destination cache entries even when the registry does not list them. Synthetic tests preserve both cache files and the selected account.
- Login rejects cache-age durations that cannot be represented as whole seconds. Previously `500ms` was accepted and persisted as zero, meaning unlimited. Parser, registry-preservation, and CLI tests cover rejection, unchanged zero semantics, and the valid `1000ms` form. Help and the interactive prompt state the precision.
- Both signaling transports publish their cancellation cause before notifying waiting credential requests. Tests reproduce the previous `(nil, nil)` result while notifier cleanup is blocked.

README repetition was reduced, stale optional-pack and server-matrix descriptions corrected, and existing focused guides updated. Historical records and evidence were retained. No new user guide, protocol definition, dependency, public CLI option, capture schema, or server-specific workaround was added.

## Validation

The initial quality gate and final `tools/format.ps1` plus `tools/quality.ps1 -Race` passed. Static analysis reported zero findings and the vulnerability scan found no vulnerabilities in called functions. Both local replacement modules separately passed `go test -race -count=1 ./...`. Each defect has a deterministic pre-fix failure and a post-fix check, including the embedded JavaScript regression invoked by the viewer tests. The dangling-symlink test skips on this Windows host because symlink creation is unavailable; Linux CI must exercise it. Focused CLI help tests and local documentation link checks passed.

Browser validation must use the clean committed build and cover live closure, filters, details, ordinary blob previews, empty blobs, and absent blobs. The release matrix must use the same stamped runtime for all six servers, retain at least 30 seconds after final spawn, and follow Enchanted's current OP SMP form and transfer. The maintainer accepted CubeCraft rejection, kick, or ban risk for this task. This does not imply a server policy exception.

## Compatibility and limits

BedrockPackArchive's importer still receives the same valid v1 contract. Consumer commit `2e5c9e2fcbf21812e477198760ffcf994144dbb6` adds the paired synthetic compatibility check. Full consumer tests against the working-source producer verifier, vet, build, and whitespace checks pass. The check rejects inconsistent layout without changing archive bytes, manifests, per-file-key counts, or index provenance. Both matched commit SHAs and the stamped-runtime result belong in the release audit.

The account store has no new cross-process transaction or locking guarantee. The signaling tests exercise local shutdown rather than forcing a public service failure. Automated tests and CLI sessions do not establish retail rendering or gameplay. The review does not claim that all possible defects or performance bottlenecks have been eliminated.
