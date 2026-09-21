# Audit AUDIT-0127 - Cancel early-login work with the downstream connection

- Commit subject: `Cancel early-login work with the downstream connection`
- Change type: `runtime | test | documentation`
- Runtime impact: `yes, cancellation only`

## Intent

Keep upstream login and URL resource-pack preparation tied to the physical
downstream connection even before the observing network wrapper returns from
`Listener.Accept`.

## Scope

- Add a lifecycle-context helper that prefers the tracked observing context and
  falls back to the listener connection context during early login.
- Use the helper when starting the upstream dial and URL resource-pack cache.
- Add deterministic tests for both context sources.
- Document the listener and wrapper call-order boundary in the resource-pack
  research and user documentation.

## Evidence and provenance

The gophertunnel listener creates and configures its `Conn` before running the
login callbacks. The observing `Network.wrap` calls `ConnectionContextFunc`
only when the completed listener connection is accepted. The listener `Conn`
already exposes the transport cancellation context through `Context()`. The
change preserves packet ordering, resource-pack selection, and raw evidence;
it only changes cancellation ownership for work already scoped to that client.

## Validation

- Focused tests: `go test ./internal/proxy -run 'Test(ConnectionLifecycleContext|LinkedConnectionContext)' -count=1 -v` passed.
- Full quality gate: `tools/quality.ps1 -LowMemory` and `git diff --check` passed before commit.
- Live validation: not observed. The user had quit the local server session, and
  no public server or external account was contacted.

## Limitations and follow-up

The tests prove context selection but do not measure cancellation latency during
a real client resource-pack download. A future owner-controlled local capture
can verify the resulting `upstream.dial_error` or URL-cache cancellation event.
