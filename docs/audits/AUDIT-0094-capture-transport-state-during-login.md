# Audit AUDIT-0094 - Capture transport state during login

- Commit subject: `Capture transport state during login`
- Change type: `capture-observability`
- Runtime impact: `none on forwarded wire behavior`

## Intent

Make compression and encryption state available on transport payload events from the beginning of each Minecraft connection, including the login sequence.

## Scope

- Add opt-in connection callbacks to the vendored gophertunnel listener and dialer.
- Associate downstream and upstream `*minecraft.Conn` values with their hop before the first login packet.
- Extend the proxy loopback assertion so both channels expose negotiated compression before `session.negotiated`.
- Update capture research, fix history, changelog, and vendored patch provenance.

## Evidence and provenance

- `internal/bedrock/network.go` records transport payloads before gophertunnel returns a completed login connection to the proxy.
- `third_party/gophertunnel/minecraft/listener.go` and `dial.go` configure the connection before starting login, making an opt-in callback a narrow observation hook.
- No external source was copied and no public server was contacted.

## Automated validation

- `go test ./internal/proxy -run TestRunnerForwardsBidirectionalPacketsAndCapturesSession -count=1` passed, including the new pre-`session.negotiated` assertions.
- `tools/quality.ps1` passed with `GOMAXPROCS=2` and `GOFLAGS=-p=1`. It completed gofmt, PowerShell and JavaScript syntax checks, module checks, license inventory, root tests, build, golangci-lint, govulncheck, artifact checks, and whitespace checks.
- The nested gophertunnel module test suite passed with `go test ./...`.
- The focused proxy integration test and `git diff --check` passed.

## Live validation status

`not_run`. Direct vanilla-client validation remains pending in an owner-controlled local environment.

## Known limitations and follow-up

The first payload may correctly report `not_negotiated` before NetworkSettings is processed. This remains an application-batch observation boundary and does not expose raw UDP or RakNet framing.
