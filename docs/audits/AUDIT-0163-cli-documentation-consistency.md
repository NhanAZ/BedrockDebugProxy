# Audit AUDIT-0163 - Require CLI and documentation consistency checks

- Commit subject: `Require CLI and documentation consistency checks`
- Intent: prevent command help, examples, and focused documentation from drifting away from the public CLI interface.
- Runtime impact: none. This change updates agent workflow guidance only.

## Scope

- Add an explicit CLI and documentation consistency section to `AGENTS.md`.
- Require synchronized root help, `help COMMAND`, relevant docs, and focused usage tests for public CLI changes.
- Clarify when user-visible source comments require corresponding documentation updates.
- Record the maintenance rule in the fix history.

## Evidence and provenance

- The rule reflects the existing grouped help and `help COMMAND` interface introduced in commit `a20ed7b`.
- No source code, binary, authentication cache, capture, or protocol behavior was changed.

## Automated validation

- `git diff --check`
- Documentation-only review of the updated guide and fix-history entry.

## Live validation status

Not applicable. No runtime behavior changed.

## Known limitations and follow-up

- The rule relies on the normal review and test workflow to identify stale examples in files outside the focused documentation search.
