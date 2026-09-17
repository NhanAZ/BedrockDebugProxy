# AUDIT-0032 - Embed licensed viewer fonts

- Commit subject: `style: use licensed pixel fonts in viewer`
- Change type: `runtime | assets | documentation`
- Runtime impact: `yes - optional viewer presentation and embedded asset size only`

## Intent

Give the local capture viewer a consistent Minecraft-inspired visual identity while using font files that explicitly permit embedding and redistribution. Avoid distributing the proprietary Mojangles and Minecraft Ten files supplied for evaluation.

## Scope

Embed the unmodified `Minecraft-Bold.otf` file for uppercase headings and the unmodified `Monocraft.ttf` file for every other interface and packet-data element. Disable programming ligatures so packet text remains visually literal. Serve both files from the existing token-protected loopback viewer, keep the content security policy self-only, expose the OFL text from the footer, include it in release license bundles, and record immutable source revisions and hashes.

Packet forwarding, capture recording, capture schema, authentication, protocol handling, analysis results, and the viewer API are unchanged. The viewer contacts no font CDN or other external service.

## Evidence and provenance

- `IdreesInc/Monocraft` commit `e498bf70aeb25b4bdcff1e44d878fb2cb4f7c2a9` distributes `dist/Monocraft-ttf/Monocraft.ttf` under SIL Open Font License 1.1 and describes it as a fan-made implementation containing no original game font or asset files.
- `IdreesInc/Minecraft-Font` commit `261ac77fbf28796ca09c22eb83ecdfe386c4b838` distributes `Minecraft-Bold.otf` under SIL Open Font License 1.1 and describes it as a fan-made implementation containing no original game font or asset files.
- The exact file hashes, immutable source links, attribution, and license boundary are recorded in `THIRD_PARTY_NOTICES.md`.
- The proprietary files at `[local path]` and `[local path]` were inspected but were not copied into the repository or executable.

## Validation

- Automated checks: `pass - tools/format.ps1, focused internal/viewer tests, embedded font SHA-256 assertions, third-party license inventory and bundle generation, tools/quality.ps1, and git diff --check`
- Live validation: `not_required - presentation-only viewer change with no network or capture behavior impact`
- Capture or report: `synthetic viewer HTTP asset tests and a local closed-capture browser check at 1280 px confirming both font loads and no page-level horizontal overflow`

## Limitations and follow-up

The heading font covers a smaller Latin character set than Monocraft and is intentionally limited to fixed English interface headings. Captured packet data always uses Monocraft. Browser font rendering can vary slightly by platform. These upstream fan projects are not affiliated with or endorsed by Mojang or Microsoft.
