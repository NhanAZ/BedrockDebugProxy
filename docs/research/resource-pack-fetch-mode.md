# Resource-pack-only capture mode

## Decision

Implement `run --packs-only` as a mode of the existing runner rather than as a separate downloader command or offline archive fetcher. It uses the same configured upstream dial, transport observation, URL-pack adapter, canonical capture recorder, automatic artifact writer, and optional in-session decryption as a normal run. It closes the dialled connection after the resource-pack exchange and never calls `Conn.DoSpawn`.

The operator receives the login and resource-pack evidence from one live upstream connection. The feature does not promise that a server will treat the connection as invisible or exclude it from connection, account, or session logs. It does not retrieve an arbitrary URL or reuse an archive or content key from a previous session.

## Source review

- Sandertv gophertunnel `v1.61.0`, commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`. [`minecraft/dial.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/dial.go) and [`minecraft/conn.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/conn.go) were reviewed as the API baseline. The BedrockDebugProxy fork's `DialContextNetwork` waits for its `connected` signal, and the client-side connection marks that state after the world-start sequence. Therefore, the separate `Conn.DoSpawn` API alone does not provide a pre-spawn return boundary in this fork. The local fork adds the opt-in `Dialer.ResourcePackExchangeOnly` signal after accepting `ResourcePackStack` and flushing `PackResponseCompleted`; default dial behavior is unchanged. The loopback regression test proves the runner returns and verifies a pack archive when the test server never sends `StartGame`.
- bedrock-tool/bedrocktool, commit `85d5cfe1545c8d853be2859144a8357539ffc0f2`, [`subcommands/resourcepack-d.go`](https://github.com/bedrock-tool/bedrocktool/blob/85d5cfe1545c8d853be2859144a8357539ffc0f2/subcommands/resourcepack-d.go). This was reviewed as a product and behavior reference for retrieving offered packs. BedrockDebugProxy does not copy its code or create a separate key store. Captures remain in the canonical local evidence store, and content keys remain sensitive.
- NhanAZ-Tools/EasySpecter, maintainer-authorized private revision `6939f8aa1ad70fc694d370e3b4b33e6d3e49482c`, files `specter/xbox.go` and `normal.go`. The reviewed code uses gophertunnel's authenticated dial path directly. The source was used only to understand the headless connection lifecycle. No code was copied or adapted.
- The public [`Phuongaz/EasySpecter`](https://github.com/Phuongaz/EasySpecter) fork was identified as an older reference, but the maintainer-owned private revision above was the reviewed version for this decision.

Accessed 2026-09-23. Repository license information was reviewed for source-reuse triage. No source from these projects was copied, so this feature adds no third-party code provenance obligation.

## Capture and privacy boundaries

The mode captures transport and packet events from the live upstream connection and records downloaded pack archives with the existing `resource_pack.archive` event. Existing raw `ResourcePacksInfo`, `ResourcePackStack`, packet ordering, and connection metadata remain available. `operation_mode` is set to `packs_only`, and a `resource_pack.fetch_complete` event marks successful completion and states that no world spawn was requested. Capture metadata records the account label, not an authentication token.

No local Minecraft client supplies `ClientData` in this mode. The dialer therefore uses gophertunnel's default client metadata, which currently selects Android defaults. The structured `session.connection_metadata` event and the `client_metadata_source` capture value make this explicit. A server may vary its pack offer by client platform or account state, so a packs-only capture is not evidence that every client profile would receive the same set.

Server-supplied content keys and pack bytes stay in the local capture. Decryption remains opt-in through `--decrypt-resource-packs` and uses only the content key on the current connection. The mode has the existing bounded fifteen-minute upstream login and pack-exchange deadline. Servers that defer packs until later gameplay are outside its coverage.

## Validation status

The implementation has not been validated against a public server. The deterministic loopback tests `TestRunPacksOnlyCapturesUpstreamWithoutListener` and `TestDialerResourcePackExchangeOnlyStopsBeforeWorldStartup` exercise an authentication-disabled local Bedrock listener, an offered pack archive, and `StartGame` sent immediately after the pack-only boundary. They verify successful completion, exact archive SHA-256, capture integrity, and no client world-spawn packets. The runner test also passes an invalid client-listener address to `proxy.New`, demonstrating that packs-only does not require the proxy's client-facing listener. The full project quality gate must pass on the final diff before commit. These local fixtures do not establish which packs a real server offers for another account or client profile. Public-server testing is not implied by this feature or this note.
