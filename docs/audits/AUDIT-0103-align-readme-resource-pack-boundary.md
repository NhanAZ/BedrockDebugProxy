# Audit 0103 - Align README resource-pack boundary

- Intent: align the README status section with the resource-pack offer metadata behavior implemented in the current runtime.
- Scope: documentation only. Updated the README known-boundary entry for URL-advertised resource packs.
- Affected interfaces: `README.md` user-facing behavior and limitation description. No Go code, capture schema, protocol behavior, or binary output changed.
- Evidence and provenance: `internal/proxy/proxy.go`, `docs/resource-packs.md`, and `docs/fix-history.md` FIX-0076 describe retained offer metadata, proxy-served URL archives, and observable retrieval failures. The previous README sentence was stale relative to those sources.
- Automated validation: run the low-memory quality gate, `git diff --check`, and the documentation consistency test before commit.
- Live validation: not applicable. This commit has no runtime impact.
- Known limitations: mixed URL and RakNet resource-pack behavior still needs an owner-controlled real-client validation as recorded in FIX-0076.
- Follow-up: keep the README boundary text synchronized when resource-pack delivery behavior changes.
