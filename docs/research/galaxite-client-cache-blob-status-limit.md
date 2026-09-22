# Galaxite ClientCacheBlobStatus decode limit

## Observation

The Galaxite release session on 2026-09-22 used stamped runtime revision `93435c5119e2899dc1958d8cec0fa63447bd611c` and closed with 69,141 events. The capture verifier found one blocking decode error at sequence 51,524:

```text
decode packet *packet.ClientCacheBlobStatus: slice length was too long: length of 1302 (max 1024)
```

The raw packet was retained. The normal session still connected, spawned, exchanged traffic in both directions, and completed artifact generation, but the release report is `fail` because the decode error gate is not satisfied.

## Source comparison

The vendored gophertunnel revision uses a generic `maxSliceLength = 1024` in `protocol.SliceOfLen` and `protocol.FuncSliceOfLen`. `ClientCacheBlobStatus.Marshal` uses those helpers for both hash lists, so a valid list with more than 1,024 entries becomes a decode error before forwarding can expose a typed view.

The matching BedrockProtocol implementation reads the miss and hit counts as VarInts and iterates each declared count without a 1,024-element protocol limit. These sources support classifying the observed failure as a likely integration or dependency safety-limit mismatch, not proof of malformed Galaxite input.

- Vendored source: `third_party/gophertunnel/minecraft/protocol/io.go` and `third_party/gophertunnel/minecraft/protocol/packet/client_cache_blob_status.go`.
- Gophertunnel master at revision [`7a556a07335b663744b50d38062636ad8283f314`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/protocol/io.go#L120-L145), accessed 2026-09-22.
- BedrockProtocol master at revision [`8eb5a390a8d913ee54b97ac71f42899ecbb5c419`](https://github.com/pmmp/BedrockProtocol/blob/8eb5a390a8d913ee54b97ac71f42899ecbb5c419/src/ClientCacheBlobStatusPacket.php#L53-L71), accessed 2026-09-22.

## Status and validation

The decoder now gives only `ClientCacheBlobStatus` a bounded 16,384-entry limit. Ordinary protocol slices keep the upstream 1,024-entry bound. `TestClientCacheBlobStatusAllowsObservedHashCount` encodes and decodes 1,302 miss hashes with the strict reader enabled and checks both list boundaries. The Galaxite release report remains `fail` until a new stamped runtime revision passes without a blocking decode error.
