# Validation artifacts

The ignored `validation/local/<REVISION>/` directory may contain small `bedrockdebugproxy.validation.v1` JSON reports generated from agent-controlled protocol-client sessions by `tools/create-validation-report.ps1`.

A report identifies the tested revision and records sanitized evidence for one server session. It is not a raw capture, a compatibility guarantee, or permission to publish captured content. Review every report before uploading it.

Routine runtime pull requests may use an exact-revision packs-only smoke report when it observes the affected login or resource-pack flow. Such reports are expected to say `incomplete` because packs-only mode does not request world spawn. Changes outside that boundary need focused automated, fixture, or loopback evidence. Agents may use Golem CLI for disposable local server/plugin scenarios, but its simulation journal is not a proxy capture or public-server report. Each official release requires passing agent-controlled protocol-client reports for the six servers listed in `docs/validation.md`, always with `--decrypt-resource-packs` and successful decryption of every encrypted offered pack. The release workflow does not add a separate repeated CubeCraft rejection, kick, or ban acknowledgement; it still uses the normal read-only path and does not imply server permission. Do not record or publish private server discussions.

The Enchanted server sends an in-game form when the player joins. The release report must select any minigame from that form and observe the resulting transfer and next-hop resource-pack flow. A hub-only connection does not pass the release gate.

Upload sanitized reports as pull request or release artifacts without committing them into the tested candidate. A report commit would create a new head that the stamped binary did not test. Long-term release evidence belongs with the immutable release assets. Never place packet payloads, addresses, credentials, identifiers, content keys, decrypted resource packs, or third-party assets there.
