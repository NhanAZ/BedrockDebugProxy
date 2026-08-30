# Validation artifacts

The ignored `validation/local/<REVISION>/` directory may contain small `bedrockdebugproxy.validation.v1` JSON reports generated from real sessions by `tools/create-validation-report.ps1`.

A report identifies the tested revision and records sanitized evidence for one server session. It is not a raw capture, a compatibility guarantee, or permission to publish captured content. Review every report before uploading it.

Runtime-affecting pull requests require a passing The Hive report for the exact revision. Flow-specific changes require an additional report that observes the changed flow. Each official release requires reports for the six servers listed in `docs/validation.md`.

Upload sanitized reports as pull request or release artifacts without committing them into the tested candidate. A report commit would create a new head that the stamped binary did not test. Long-term release evidence belongs with the immutable release assets. Never place packet payloads, addresses, credentials, identifiers, content keys, decrypted resource packs, or third-party assets there.
