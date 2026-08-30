# Validation artifacts

This directory may contain small `bedrockdebugproxy.validation.v1` JSON reports generated from real sessions by `tools/create-validation-report.ps1`.

A report identifies the tested revision and records sanitized evidence for one server session. It is not a raw capture, a compatibility guarantee, or permission to publish captured content. Review every report before committing it. Never place packet payloads, addresses, credentials, identifiers, content keys, decrypted resource packs, or third-party assets here.

Runtime-affecting pull requests require a passing The Hive report for the exact revision. Flow-specific changes require an additional report that observes the changed flow. Each official release requires reports for the six servers listed in `docs/validation.md`.
