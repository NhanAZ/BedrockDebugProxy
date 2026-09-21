# Audit 0128 - Strict resource-pack offer decoding

- Intent: prevent the URL resource-pack adapter from acting on a partially decoded `ResourcePacksInfo` body.
- Scope: `internal/proxy/resource_pack_url.go`, its regression test, resource-pack documentation, fix history, and the Unreleased changelog.
- Behavior: URL prefetch now requires a complete typed offer with no unread trailing bytes. The original raw packet remains the canonical evidence and is not changed by this auxiliary check.
- Evidence and provenance: the boundary follows `decodeRequestChunkRadius` and the protocol reader used by the existing `ResourcePacksInfo` observer. The test appends one trailing byte to a valid empty offer.
- Automated validation: focused proxy tests and the project low-memory quality gate are required before commit. `git diff --check` is required.
- Live validation: not run. No public server or external account was contacted.
- Known limitations: decode failures remain dependent on the canonical packet observer's capture reporting. This change only prevents URL-prefetch side effects from an incomplete offer.
- Follow-up: a local loopback test may later assert that an incomplete URL offer does not reach an HTTP fixture, if future decoder paths add auxiliary work.
