# Empty final resource-pack URL

## Finding

This is a local integration regression in FIX-0075, not evidence that a server sent each pack twice. At pre-fix revision dcaeba5bfa1712f2c2ca09b30d2feae326feb2db, decodeResourcePacksInfo uses bytes.NewReader. The vendored protocol.Reader.String consumes a length and then calls Read even when that length is zero. For an empty final DownloadURL, bytes.Reader returns EOF. Observe leaves offerOrder empty. Normal chunk downloads still populate Conn.ResourcePacks and the per-connection cache. Merge then appends both lists.

## Primary sources

- Go 1.26.6 [bytes.Reader.Read](https://github.com/golang/go/blob/go1.26.6/src/bytes/reader.go) checks exhaustion before copying, including for a zero-length destination.
- Go 1.26.6 [bytes.Buffer.Read](https://github.com/golang/go/blob/go1.26.6/src/bytes/buffer.go) accepts a zero-length read at EOF.
- The in-tree gophertunnel protocol/reader.go String method and minecraft/packet.go parseData path establish the string read and canonical bytes.Buffer input. Its upstream baseline is v1.61.0 with separately documented MIT patches. No dependency source is copied or changed here.
- internal/proxy/resource_pack_url.go decodeResourcePacksInfo, Observe, Store, and Merge establish the local integration. Earlier mixed-offer coverage ends with a non-empty URL. Earlier packs-only coverage accepted any positive archive count.

The exact Go URLs and method bodies were inspected on 2026-10-05. They are behavioral references rather than imported source.

## Correction and validation

Use bytes.NewBuffer in the auxiliary decoder, matching the canonical decoder. Preserve panic recovery and the unread-byte check. Do not change packet definitions or filter archive events to hide the error.

TestDecodeResourcePacksInfoAcceptsFinalEmptyDownloadURL verifies the entry and unchanged input bytes. TestURLResourcePackCacheMergePreservesChunkOnlyOfferOrder verifies two cached chunk downloads appear once in server order despite reversed completion order. Both tests fail before the fix. The loopback packs-only test now requires exactly one archive and a count of one. Malformed, trailing-body, mixed-offer, and URL fallback coverage remains applicable.

BedrockPackArchive's matched fixture adds an empty-URL offer at initial and transferred hops. It requires one reconstructed pack and index entry with unchanged archive hash, root and nested manifests, plaintext contents.json, synthetic key count, and source sequence. No importer production change is needed because the capture contract is unchanged.

## Archival boundary

A display name is not an archive identity. Distinct UUIDs can share a name and a changed pack can retain its version string. Capture preserves the exact encrypted archive before deriving the plaintext archive and contents.json. Asset intake must preserve those reconstructed bytes without pretty-printing JSON. For a maintainer-selected latest-offer-only main, old variants and historical formatting stay in their immutable snapshots.

Synthetic evidence establishes the root cause. A stamped The Hive packs-only report must confirm the affected login and pack flow. It does not establish retail-client spawn or gameplay.
