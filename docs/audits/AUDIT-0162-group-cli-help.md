# Audit AUDIT-0162 - Group and discover CLI help

- Commit subject: `Group and discover CLI help`
- Intent: make the command-line workflow discoverable without requiring users to memorize a flat list of commands.
- Runtime impact: no packet, authentication, networking, or capture behavior changes. Root help and the new `help COMMAND` route only format guidance.

## Scope

- Group root help into proxy, authentication accounts, capture inspection, and information sections.
- Add concise command-specific help for every command.
- Document the help workflow in the README and authentication guide.

## Evidence and provenance

- Usage strings reuse the existing command syntax and flags. No command implementation or storage path was changed.
- Help output is static and does not read credentials, captures, or network state.

## Automated validation

- `go test ./cmd/bedrock-debug-proxy`
- `git diff --check`
- Focused tests cover grouped help, command-specific help, and unknown command handling.

## Live validation status

Not applicable. This is a read-only CLI presentation change with no runtime protocol impact.

## Known limitations and follow-up

- The root help intentionally summarizes common flags. Use each command's flag parser for the complete option list.
