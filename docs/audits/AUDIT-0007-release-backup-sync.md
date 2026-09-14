# Audit 0007 - Synchronize releases with the private source backup

- Commit subject: `Require release synchronization with the private source backup`
- Change type: Documentation and release process
- Runtime impact: None. No product binary, protocol behavior, or capture schema changed.
- Intent: Ensure every verified release is followed by a checked synchronization of Git-tracked source history, branches, and tags to `NhanAZ/BedrockDebugProxy-Backup`.
- Scope: `AGENTS.md`, `docs/releasing.md`, and this audit record.
- Root cause or reason: The backup repository existed and contained the current `main` history, but the release workflow did not require agents to update or verify it after publishing. This is a process gap, not a runtime defect.
- Evidence: The configured local `backup` remote points to `https://github.com/NhanAZ/BedrockDebugProxy-Backup.git`, and the repository is private. The documented commands use `git push --all` and `git push --tags`, then verify the tested `main` revision and release tag.
- Provenance: No external source code was copied. The backup destination is project-owner controlled.
- Automated validation: Run `git diff --check` and `./tools/quality.ps1` before committing. Review the rendered Markdown and verify the PowerShell examples are syntactically consistent with the existing release workflow.
- Live validation: Not applicable. The change affects release procedure only. The next release must exercise the backup push and ref verification.
- Known limitations: GitHub Release assets are not mirrored by Git pushes. Ignored captures, authentication state, resource-pack keys, private keys, and build output remain outside the backup. The backup repository must remain private and accessible to the release agent.
- Security and compatibility impact: This change adds no credentials or sensitive artifacts. It explicitly keeps recovery source history separate from secrets and local captures.
- Follow-up: At the next release, record the backup remote and verified refs in the release handoff without publishing sensitive repository contents.
