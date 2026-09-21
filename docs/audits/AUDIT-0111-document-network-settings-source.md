# AUDIT-0111 - Document the local NetworkSettings source

- Commit subject: `Document the local NetworkSettings source`
- Change type: `documentation`
- Runtime impact: `none`

## Intent

Record the primary PocketMine and vendored listener sources behind the observed local NetworkSettings threshold
difference, without turning one local server setting into a generic normalization rule.

## Scope

Update `docs/research/network-settings-parity.md` with the PocketMine `COMPRESS_EVERYTHING` source, the gophertunnel
listener default, and the handshake timing boundary. No runtime code, capture schema, or compression behavior changes.

## Evidence and provenance

The controlled capture records upstream threshold `1` and downstream threshold `256`. PocketMine's
`src/network/mcpe/handler/SessionStartPacketHandler.php` sends `NetworkSettingsPacket::COMPRESS_EVERYTHING`, while
the vendored listener applies its documented `256` default. The source paths are local owner-controlled evidence.

## Automated validation

- `git diff --check`
- No runtime tests are changed because this commit only clarifies existing research evidence.

## Live validation status

Not applicable. No server or client was started and no public endpoint was contacted.

## Limitations and follow-up

The note still leaves generic threshold selection unresolved. Any runtime change requires a safe handshake design and
fresh local parity evidence across more than one server configuration.
