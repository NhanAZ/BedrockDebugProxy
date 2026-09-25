# AUDIT-0064 - Add public Hive ban and proxy reports

- Commit subject: `docs: add public Hive enforcement reports`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Add the public Hive forum and Reddit reports found during follow-up research so the incident record includes community evidence without presenting anecdotes as a diagnosis.

## Scope

Extend the Hive research note with reports of `Unfair Advantage` false positives, lag-related punishments, and VPN/proxy misclassification. Link the evidence from `docs/fix-history.md`. No source, protocol model, capture, credential, or runtime behavior changes.

## Evidence and provenance

Sources are public Hive forum threads and one Reddit thread. The note records the authors' claims as anecdotal, includes a Hive staff statement that anti-cheat details are not disclosed, and explicitly records that no indexed result names BedrockDebugProxy, `SrDuck84`, or this incident's identity pattern.

## Validation

- Automated checks: `git diff --check`, targeted documentation searches, and the existing documentation quality gate.
- Live validation: `not_observed`. This documentation change does not authorize a new Hive connection.
- Capture or report: existing Hive captures are archived in private repositories and remain sensitive.

## Limitations and follow-up

Community reports can demonstrate recurring user-visible symptoms but cannot identify Hive's private detector or establish that the reports share a cause with this incident. No runtime change is proposed.

No runtime impact.
