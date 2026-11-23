# AUDIT-0277 - Resolve repository before building protocol issue links

- Change type: `tooling | test | documentation`
- Runtime impact: `no`

## Intent

Generate a valid protocol-update guide link in issues created by the scheduled watcher.

## Root cause

The watcher built the issue body before resolving `GITHUB_REPOSITORY`. Workflow run [45215096393](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/45215096393) created issue [#12](https://github.com/NhanAZ/BedrockDebugProxy/issues/12) with a link missing the owner and repository path.

## Scope

Resolve and validate the repository before building the issue body. The watcher test now checks the generated process-guide URL. Issue #12 was updated to use the correct repository link.

## Validation

- Automated checks: `go test -run TestMojangProtocolWatcherReleaseMetadata -count=1 ./cmd/bedrock-debug-proxy` passed. `git diff --check` passed.
- Live validation: no Minecraft server or client was contacted. Issue #12 was edited to correct the generated link.
- Capture or report: none.

## Limitations and follow-up

This corrects the issue link only. It does not change release detection, protocol baselines, or issue deduplication.
