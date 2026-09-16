# AUDIT-0008 - Bedrock 1.26.50 protocol update

- Commit subject: `Update Bedrock protocol to 1.26.50`
- Change type: `protocol | runtime | documentation | tooling`
- Runtime impact: `yes`

## Intent

Update the selected Bedrock protocol from 1.26.45 protocol 2169 to 1.26.50 protocol 2193 without discarding stable gophertunnel fixes or BedrockDebugProxy's documented compatibility and observation hooks.

## Scope

The change updates the in-tree gophertunnel protocol metadata, packet IDs and pools, packet and value serializers, readers and writers, NetherNet address behavior, and its related dependencies. It adds focused project-level protocol tests and updates protocol workflow, provenance, current-status, resource-pack, legal, ecosystem, and fix-history documentation.

The capture schema, raw observation boundary, unknown-packet representation, authentication policy, transfer policy, and existing local gophertunnel and go-raknet compatibility hooks are intentionally unchanged.

## Evidence and provenance

- Mojang `bedrock-protocol-docs` release `v1.26.50` at `c0bd91f7d896cec780f1185cc548b5e46a46f5d5` confirms protocol 2193 and supplies the primary packet schemas.
- Sandertv gophertunnel `v1.61.0` at `283a5a97dfe65da94bcc0b401807f6aefa9e72ee` is the stable MIT-licensed base.
- Reviewed upstream 1.26.50 changes run through `481f3bd138766304a73d7a0412a47a87acec15ed`; the required sound-data correction comes from master through `b8bd7357c24fcba3e32f940d1afb1d3b4e43b96d`.
- Cloudburst Protocol listed 2192 for the same game version during review. The project selects 2193 because Mojang's released schema is primary and the later gophertunnel correction independently agrees.
- Mojang source was used only as specification evidence. No code was copied from Mojang, Cloudburst, PrismarineJS, Endstone, axolotl-pm, Altay, or BetterAltay.
- Detailed claims, links, and the resolved disagreement are recorded in [`docs/research/protocol-1.26.50.md`](../research/protocol-1.26.50.md).

## Validation

- Automated checks: `pass - tools/format.ps1; tools/quality.ps1; nested gophertunnel go test ./...; focused protocol metadata, registration, and encoding tests; bedrock-debug-proxy version; git diff --check`
- Live validation: `pending`
- Capture or report: `none`

## Limitations and follow-up

There is no stable gophertunnel tag for Bedrock 1.26.50 at the time of this update, so the exact composite revisions remain part of the maintained provenance. Other implementations may lag or implement different valid server packet orderings. The focused tests do not replace live interoperability evidence. Build a stamped binary from the committed revision and complete the required The Hive report before treating the runtime update as complete.
