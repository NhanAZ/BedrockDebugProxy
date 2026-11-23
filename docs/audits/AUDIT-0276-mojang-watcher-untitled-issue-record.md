# AUDIT-0276 - Handle untitled protocol issue records

- Change type: `tooling | test | documentation`
- Runtime impact: `no`

## Intent

Keep the scheduled Mojang protocol notification workflow from failing when its issue-list response contains a record without a `title` field.

## Root cause

The duplicate check accessed `.title` directly under PowerShell strict mode. GitHub Actions run [45200163935](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/45200163935) failed with "The property 'title' cannot be found on this object."

## Scope

The watcher now checks the issue record's `title` property before comparing it. It warns and skips records without a title. Pull requests remain excluded from issue matching. A synthetic case covers the malformed list entry in the existing watcher test.

## Validation

- Automated checks: `go test -run TestMojangProtocolWatcherReleaseMetadata -count=1 ./cmd/bedrock-debug-proxy` passed, including the synthetic untitled-record case. `git diff --check` passed.
- Live validation: no Minecraft server or client was contacted. No live GitHub issue was created during local validation.
- Capture or report: none.

## Limitations and follow-up

The guard handles issue-list records that lack a title. It does not change protocol release selection, the project baseline, or issue content.
