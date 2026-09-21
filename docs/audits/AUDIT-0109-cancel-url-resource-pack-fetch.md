# AUDIT-0109 - Cancel URL resource-pack fetches with the connection

- Commit subject: `Cancel URL resource-pack fetches with the connection`
- Change type: `runtime`
- Runtime impact: `yes, URL resource-pack prefetch cancellation`

## Intent

Stop in-flight HTTP resource-pack downloads when the connection context ends, so a closed downstream session does not
keep CDN and temporary-archive work running until the outer login timeout.

## Scope

Add the context-aware `resource.ReadURLLimitedContext` helper and pass the URL cache's per-connection context through
the existing bounded reader. Request creation and body reads now observe cancellation. Resource-pack offer ordering,
raw packet retention, size bounds, UUID/version/size checks, and normal RakNet fallback are unchanged.

## Evidence and provenance

`newURLResourcePackCache` already retained a connection context and checked it before and during cache access, but the
vendored URL reader used `http.Get`, so a started fetch could not observe cancellation. The change is a narrow local
patch to the in-tree Sandertv gophertunnel copy. No external source code was copied.

## Automated validation

- `go test ./internal/proxy -run 'TestURLResourcePackCache' -count=3`
- The new loopback fixture holds an HTTP response open, cancels the cache context, and confirms the worker exits with
  a context-canceled error and returns no pack.
- Full `tools/quality.ps1 -LowMemory`, formatting, build, vulnerability scan, and `git diff --check` are required
  before commit.

## Live validation status

Pending. The change only affects cancellation of URL-prefetched resource packs. A stamped exact-revision The Hive
baseline and owner-controlled HTTP resource-pack flow remain required by the project validation policy.

## Limitations and follow-up

Cancellation depends on the HTTP transport honoring request contexts and does not interrupt already completed archive
parsing. The normal RakNet resource-pack path is unchanged. Real-client URL and mixed-offer validation remain pending.
