# Validation artifacts

The ignored `validation/local/<REVISION>/` directory may contain small `bedrockdebugproxy.validation.v1` JSON reports generated from real sessions by `tools/create-validation-report.ps1`.

A report identifies the tested revision and records sanitized evidence for one server session. It is not a raw capture, a compatibility guarantee, or permission to publish captured content. Review every report before uploading it.

Routine runtime pull requests may use an exact-revision packs-only smoke report when it observes the affected login or resource-pack flow. Such reports are expected to say `incomplete` because packs-only mode does not request world spawn. Changes outside that boundary need focused automated, fixture, or loopback evidence. Each official release requires passing real-client reports for the six servers listed in `docs/validation.md`. The CubeCraft release report also requires the maintainer to accept the connection risk for the exact diagnostic session. Do not record or publish private server discussions. Without the acknowledgement, use `incomplete` and keep the release gate blocked.

The Enchanted server sends an in-game form when the player joins. The release report must select any minigame from that form and observe the resulting transfer and next-hop resource-pack flow. A hub-only connection does not pass the release gate.

Upload sanitized reports as pull request or release artifacts without committing them into the tested candidate. A report commit would create a new head that the stamped binary did not test. Long-term release evidence belongs with the immutable release assets. Never place packet payloads, addresses, credentials, identifiers, content keys, decrypted resource packs, or third-party assets there.
