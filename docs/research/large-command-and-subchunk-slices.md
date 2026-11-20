# Large command and sub-chunk packet slices

## Finding

The strict gophertunnel reader applies a generic 1,024-item ceiling to protocol slices. Mojang's protocol 2225 schemas describe `AvailableCommands` collections and the standard and extra `UpdateSubChunkBlocks` collections as arrays without a `maxItems` limit. The command schema does explicitly cap chained-subcommand data at 16, which shows that the schema records a bound where one is defined.

The generic decoder ceiling is an implementation safety limit, not a protocol field limit. A declared count still has to fit the remaining packet payload. The local packet codecs therefore use a finite 65,536-item limit only for the `AvailableCommands` and `UpdateSubChunkBlocks` collections. Other protocol slices keep the 1,024-item default.

## Sources

- Mojang `bedrock-protocol-docs` at revision [`f8d1496b5efc95105bd9ae1f2f2f076a43f71d82`: `AvailableCommandsPacketPayload.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82/json/AvailableCommandsPacketPayload.json), protocol 2225. The array fields have no `maxItems` except the chained-subcommand data bound of 16.
- Mojang `bedrock-protocol-docs` at the same revision: [`UpdateSubChunkBlocksPacketPayload.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82/json/UpdateSubChunkBlocksPacketPayload.json) and [`UpdateSubChunkBlocksChangedInfo.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82/json/UpdateSubChunkBlocksChangedInfo.json), protocol 2225. The standard and extra block-change arrays have no `maxItems` limit.
- Local gophertunnel reader and packet codecs: `third_party/gophertunnel/minecraft/protocol/reader.go`, `io.go`, `packet/available_commands.go`, and `packet/update_sub_chunk_blocks.go`.

## Validation

Synthetic strict-reader round trips cover 5,999 command enum values and 3,623 standard plus 1,667 extra block changes. They check both collection lengths and boundary values. The tests do not contain a server capture or server content.
