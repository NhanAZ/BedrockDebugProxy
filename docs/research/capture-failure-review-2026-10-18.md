# Capture failure review on 2026-10-18

The maintainer requested a directory review with seven prioritized sessions and authorized a different cached account for fresh validation. The inventory contains 29 captures, 23 closed and six failed. The review scanned event kinds and warnings across the directory and inspected packet ordering and terminal evidence in the seven focused sessions. All seven pass the canonical integrity verifier. Passing integrity verification does not change a failed manifest into a successful session. Original captures and authentication caches remain local unless archived through the private evidence workflow.

## Focused evidence

| Private capture | Observation | Classification |
| --- | --- | --- |
| [session-20261018T154700Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T154700Z/releases/tag/capture-archive) | Minecraft authentication HTTPS request ends with TLS handshake timeout before pack negotiation. | External authentication or network failure. This does not establish a CubeCraft gameplay failure. |
| [session-20261018T160312Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T160312Z/releases/tag/capture-archive) | Hub URL cache reports invalid ZIP. Its chunk fallback is followed by an upstream resource-pack disconnect. | Failed URL delivery and rejected fallback. The original HTTP body/read result is not recoverable from this capture. |
| [session-20261018T160421Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T160421Z/releases/tag/capture-archive) | OP SMP retains its seven offered packs, then the upstream sends Disconnect before child spawn. Later startup writes encounter the closed downstream. | Upstream termination during startup. The backend reason and any local contribution remain unresolved. |
| [session-20261018T161138Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T161138Z/releases/tag/capture-archive) | Three Box PvP URL requests report remote TCP reset. Their chunk fallback is followed by an upstream resource-pack disconnect. | HTTP transport failure with rejected fallback. |
| [session-20261018T161537Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T161537Z/releases/tag/capture-archive) | Minecraft authentication HTTPS request ends with unexpected EOF before pack negotiation. | External authentication or network failure. |
| [session-20261018T161634Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T161634Z/releases/tag/capture-archive) | Hub spawns and sends Transfer. The route probe reports DNS timeout. No hop 2 upstream dial follows before close. | Incomplete transfer observation. No child dial was attempted in this capture. |
| [session-20261018T161845Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T161845Z/releases/tag/capture-archive) | The Mines route receives an offer and completes a ten-pack exchange, then the upstream explicitly refuses the selected backend. No child archive event is emitted. | Upstream backend refusal after pack negotiation. Zero retained child archives do not mean zero offered packs. |

Retained packet bodies were decoded with the matching checked-out protocol implementation and consumed without trailing bytes. Resource-pack responses, offers, stacks, and Disconnect order were reviewed. The previous early-offer queue, transfer completion race, initial ping bound, and normal peer-close fixes describe different mechanisms. Later successful sessions do not rewrite failed observations. Ordinary bridge termination and advertised-versus-delivered size warnings were kept separate from failed joins.

## Reproduced local defect

At revision `de502237693c755e1dbd449d0b783d96c06c9bd1`, retained gophertunnel `minecraft/resource/pack.go`, function `Read`, discarded both return values of `io.Copy(temp, r)` and parsed whatever bytes reached its temporary archive.

Three independent synthetic reader cases fail before the correction. An error before the archive or halfway through a ZIP becomes only an invalid-ZIP parser error. A complete ZIP followed by a reader error is accepted without error. The final case establishes a correctness problem beyond diagnostic wording. The correction preserves copy and close failures, removes the temporary file, and returns the original error through wrapping before ZIP parsing. A local HTTP regression delivers a complete synthetic ZIP in a chunked response but omits the final HTTP chunk. The URL cache must return unexpected EOF and retain no successful archive.

The Hub invalid-ZIP capture motivated this review, but it cannot prove that a swallowed reader error caused that particular response. The body might also have been malformed. The fix does not retry downloads, guess another endpoint, suppress failures, change the existing chunk fallback, or make an unavailable backend accept a connection. The existing MIT dependency license remains in place. The local correction and tests were written independently.

## Paired consumer

Successful archive bytes, capture schema, event fields, manifest selection, plaintext contents.json, per-file keys, ordering, and index fields are unchanged. BedrockPackArchive's synthetic root/nested manifest fixture exercises both hops, retains archive and metadata bytes, and now checks the 0.4.0 producer version in the imported index.

Validation preparation exposed a separate version gate accepting only 0.3.x despite the producer's released 0.4.0. The paired adapter accepts numeric 0.3.x and 0.4.x versions, including a numeric development suffix. Bare dev, unrelated major versions, and unreviewed minor versions remain rejected. This changes no importer data path or public command.

## Validation boundaries

The commit audits record automated gates, paired commit identities, the stamped producer fixture check, and live scope. The alternate cached account was selected through the existing CLI and its token refreshed normally. No credential or account identifier is committed here.

A packs-only session proves authenticated login and pack exchange, not retail spawn or gameplay. The existing Enchanted menu client can stop after a child exchange and cannot replace retail startup validation. A new CubeCraft connection is unnecessary to diagnose an authentication-service TLS failure and was not made during this review. Each linked private archive retains the complete unchanged capture directory, its status, filename, byte size, and SHA-256. Repository privacy and uploaded asset digests were verified. Original local captures remain retained.
