# Third-party notices and provenance

## Project code

Except where this file states otherwise, the source in this repository was written independently for BedrockDebugProxy and is licensed `GPL-3.0-or-later` under the root `LICENSE` file.

Reviewing another implementation, protocol definition, or packet capture does not by itself mean its source was copied. Research influences, source revisions, licenses, observations, and unresolved differences are recorded in `docs/research/`.

At the time of this notice, the repository contains no vendored, copied, or substantially adapted third-party source files. Update this section in the same commit if that changes. Identify the affected files, upstream URL and revision, upstream copyright holder, license, and whether the material was copied or adapted.

## Direct runtime dependencies

The Go module uses the following direct dependencies. Their source remains under their respective licenses and is not relicensed by BedrockDebugProxy.

| Module | Version | License | Role |
| --- | --- | --- | --- |
| [`github.com/google/uuid`](https://github.com/google/uuid) | `v1.6.0` | BSD-3-Clause | UUID parsing and values |
| [`github.com/sandertv/go-raknet`](https://github.com/Sandertv/go-raknet) | `v1.15.2-0.20260705184311-0d1fd09e2cf6` | MIT | RakNet transport |
| [`github.com/sandertv/gophertunnel`](https://github.com/Sandertv/gophertunnel) | `v1.61.0` | MIT | Bedrock sessions, protocol, authentication, and resource packs |
| [`golang.org/x/oauth2`](https://github.com/golang/oauth2) | `v0.36.0` | BSD-3-Clause | Authentication token source API |

The complete direct and transitive module graph and exact checksums are recorded in `go.mod` and `go.sum`. Distributions that include dependency source or compiled dependency code must preserve the license and notice material required by those dependencies.

## Development tools

`tools/quality.ps1` invokes pinned releases of golangci-lint and govulncheck through `go run`. They are development tools, not runtime dependencies and are not added to the BedrockDebugProxy module graph.
