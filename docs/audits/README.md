# Commit audit records

Commit audit records give an AI agent or maintainer a short, evidence-backed explanation of what a commit means without requiring it to infer intent from a diff alone.

Audit records are not the user-facing changelog. Add concise release-relevant entries to the root [`CHANGELOG.md`](../../CHANGELOG.md) for user-visible changes, while keeping technical evidence, provenance, validation, limitations, and follow-up work in the audit record.

## Policy

Every authored non-merge commit includes exactly one Markdown record whose filename contains a stable change ID such as `AUDIT-0001` in this directory. `README.md` and `template.md` are policy support files, not commit records. Documentation-only and formatting-only commits use the same format and explicitly state that runtime behavior is unchanged. Historical commits before this policy are not rewritten. Their history remains in Git and their evidence remains in the existing research notes, architecture decisions, captures, and validation reports.

The record is created before the commit and uses a stable change ID such as `AUDIT-0001`. It must not depend on the final commit hash. Git metadata is the source of truth for the hash, author, timestamp, and parent history. The commit subject should name the change, and the audit record should be staged with the files it describes.

Before starting a bug fix, search [`../fix-history.md`](../fix-history.md) and its linked decisions, research notes, captures, and validation reports. Update the matching history entry or add a new stable `FIX-NNNN` entry when the root cause is different.

## Required contents

- Intent and user-visible problem
- Scope and affected files, interfaces, or protocol boundaries
- Root cause or reason for the change, distinguishing evidence from inference
- Provenance and external sources when applicable
- Automated validation and exact commands or results
- Live validation status and capture or report paths when applicable
- Known limitations, unresolved uncertainty, and follow-up work
- Security, licensing, capture, or compatibility impact when relevant

Copy [`template.md`](template.md) for a new record. Do not put credentials, raw captures, resource-pack keys, decrypted assets, or other sensitive data in an audit record.
