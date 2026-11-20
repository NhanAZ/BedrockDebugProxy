# AUDIT-0266 - Record paired BedrockPackArchive validation

- Commit subject: `Record paired BedrockPackArchive validation`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Complete the producer audit with the exact paired BedrockPackArchive commit and the result of validating its synthetic transferred-hop importer fixture against the stamped producer binary.

## Scope and provenance

The producer change is commit `71eb362293e01abe2710837a5e1c7a8995aa388b`. The paired consumer change is BedrockPackArchive commit `731203abc1c49403493d4decccfd5b6be93bb1c8`. The fixture includes a synthetic best-effort HTTPS probe warning before transferred-hop archive events. It contains no real capture, server endpoint, account data, or third-party pack.

The exact stamped producer binary reports Bedrock 1.26.62 and protocol 2225 and has SHA-256 `35F03997C1D6A77B9A147B076B65E7C8F025F28C9DE6742AFE58DD0DB6016E5D`.

## Validation

- `BPA_TEST_BEDROCK_DEBUG_PROXY=<stamped binary> go test -count=1 -v -run '^TestImportCapturePreservesRootManifestAndPackMetadata$' .` passed in BedrockPackArchive.
- The consumer's full tests, vet, build, and whitespace check passed.
- The producer's `tools/quality.ps1` passed before commit `71eb362`.
- The authenticated Enchanted fetch reached Hub spawn and verified its canonical capture with the stamped producer. The fetch remains incomplete because `Trending Games` and `Other Games` returned no follow-up form or transfer in the observed menu wait. The repeated-menu guard stopped the `Steal a Mob` cycle. This is not the producer's six-server release gate.

See [AUDIT-0265](AUDIT-0265-transfer-probe-fallback-diagnostics.md) for the code change, paired commit identities, limitations, and live validation plan.
