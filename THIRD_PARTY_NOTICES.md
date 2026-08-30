# Third-party notices and provenance

## Project code

Except where this file states otherwise, the source in this repository was written independently for BedrockDebugProxy and is licensed `GPL-3.0-or-later` under the root `LICENSE` file.

Reviewing another implementation, protocol definition, or packet capture does not by itself mean its source was copied. Research influences, source revisions, licenses, observations, and unresolved differences are recorded in `docs/research/`.

`internal/experience/signaling.go` is substantially adapted from the legacy and JSON-RPC signaling implementations in [`NhanAZ-Tools/bedrocktool`](https://github.com/NhanAZ-Tools/bedrocktool) at commit `acbe0cf671928131112e91da982a5c42aa70b498`. The upstream work is copyright its contributors and distributed under GNU GPL version 3. The adapted file is distributed under those terms. BedrockDebugProxy as a combined work may be distributed under GNU GPL version 3, which is permitted by the project's `GPL-3.0-or-later` choice. Exact source paths, later cross-check revisions, and behavioral evidence are recorded in [`docs/research/experience-routing.md`](docs/research/experience-routing.md).

The resource-pack decryption behavior was cross-checked against MIT, Apache-2.0, LGPL-3.0, and AGPL-3.0 implementations plus official NIST AES validation material. No source from those projects was copied or adapted. Exact revisions, file links, licenses, the resolved header-field discrepancy, and validation boundaries are recorded in [`docs/research/resource-pack-encryption.md`](docs/research/resource-pack-encryption.md).

## Direct runtime dependencies

The Go module uses the following direct dependencies. Their source remains under their respective licenses and is not relicensed by BedrockDebugProxy.

| Module | Version | License | Role |
| --- | --- | --- | --- |
| [`github.com/coder/websocket`](https://github.com/coder/websocket) | `v1.8.14` | ISC | Minecraft signaling WebSocket transport |
| [`github.com/df-mc/go-nethernet`](https://github.com/df-mc/go-nethernet) | `v1.0.20` | MIT | NetherNet and WebRTC upstream transport |
| [`github.com/df-mc/go-playfab/v2`](https://github.com/df-mc/go-playfab) | `v2.0.2` | MIT | Minecraft services authentication |
| [`github.com/df-mc/go-xsapi/v2`](https://github.com/df-mc/go-xsapi) | `v2.0.3` | MIT | Xbox Live session used for service authentication |
| [`github.com/google/uuid`](https://github.com/google/uuid) | `v1.6.0` | BSD-3-Clause | UUID parsing and values |
| [`github.com/sandertv/go-raknet`](https://github.com/Sandertv/go-raknet) | `v1.15.2-0.20260705184311-0d1fd09e2cf6` | MIT | RakNet transport |
| [`github.com/sandertv/gophertunnel`](https://github.com/Sandertv/gophertunnel) | `v1.61.0` | MIT | Bedrock sessions, protocol, authentication, and resource packs |
| [`golang.org/x/oauth2`](https://github.com/golang/oauth2) | `v0.36.0` | BSD-3-Clause | Authentication token source API |

The complete direct and transitive module graph and exact checksums are recorded in `go.mod` and `go.sum`. Distributions that include dependency source or compiled dependency code must preserve the license and notice material required by those dependencies.

## Development tools

`tools/quality.ps1` invokes pinned releases of golangci-lint and govulncheck through `go run`. They are development tools, not runtime dependencies and are not added to the BedrockDebugProxy module graph.
