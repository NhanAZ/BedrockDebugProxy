# AUDIT-0212 - Correct generated login skin defaults

- Commit subject: `fix: supply valid generated login skin defaults`
- Change type: compatibility
- Runtime impact: yes

## Intent and scope

Default generated login metadata is rejected by strict skin parsers before resource-pack negotiation. Add two omitted-field defaults in the existing gophertunnel defaultClientData path. Supplied skin metadata, image bytes, dimensions, identity, and authentication behavior remain preserved. Invalid nonempty observations are not silently corrected. No protocol layout or capture field changes.

## Evidence and provenance

The original hot fixes were saved and discarded before investigation. The clean baseline reproduced an empty SkinArmSizeType rejection on the isolated Axolotl-PM runtime. [The unchanged failed capture](https://github.com/BedrockDebugProxy-Captures/session-20261014T173916Z/releases/tag/capture-archive) is private and its downloaded archive matches SHA-256 `0a4158642c5644ab6414c18fa72b15f7b90127b5fec452e88c792625f6fd10e5`. The original local capture remains retained.

[The maintenance review](../research/maintenance-2026-10-15.md) links the matching enum, color parser, and client-data conversion. Those source files were verified byte-for-byte apart from line endings against immutable BedrockProtocol revision `0e13c077d1060ae9e293908eff81fbef600ead5c`. This is an independent local correction after source study. No PHP source is copied.

## Validation

The serialized generated-login constraint test failed on the previous source and passes after the correction. Focused tests cover offline and authenticated defaults and preservation of supplied valid and invalid metadata. `tools/format.ps1` and `tools/quality.ps1 -LowMemory -Race` pass, including the full race-enabled root suite, viewer checks, static analysis, build, license inventory, zero reachable vulnerabilities, and diff checks. The replacement gophertunnel minecraft, NBT, and packet suites also pass with `go test -race -count=1`. Final stamped loopback and public-server validation follow this commit.

BedrockPackArchive `8afee20944f45fad3faa545e7cdb13d923a26108` passes its updated synthetic metadata fixture with the stamped metadata producer `27325cecaf36e9bf90a684742fc31c1973f611b6`. The same unchanged archive-contract fixture will use this committed producer binary. Archive, manifest, contents.json, declared key counts, and index provenance are covered at both initial and transferred hops. The importer requires no production change. The final paired audit records the exact producer and consumer revisions.

## Human validation

With the retail 26.52 client, join an owner-controlled Axolotl or Bedrock Dedicated Server through the proxy, accept offered packs, and observe spawn. Expect no generated skin parsing failure. Keep the closed capture and server log if it fails. Synthetic spawn and public packs-only checks do not prove retail gameplay or release readiness.
