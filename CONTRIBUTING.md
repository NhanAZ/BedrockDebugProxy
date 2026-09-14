# Contributing

This file is the shortest path from a local change to a merge-ready pull request. It links to the detailed rules only when they are needed.

## Before changing code

1. Read `AGENTS.md` for the project boundaries.
2. Trace the affected code path, its callers, existing helpers, tests, and relevant documentation.
3. Keep the change focused. Do not combine a bug fix with unrelated cleanup or protocol updates.
4. Search [`docs/fix-history.md`](docs/fix-history.md) for the reported symptom and inspect its linked evidence before coding. Classify the issue as a known fix, regression, limitation, duplicate, or new root cause.
5. For protocol work, follow [`docs/protocol-updates.md`](docs/protocol-updates.md) and record evidence before coding.

## Before opening a pull request

1. Format and run the complete automated gate.

   ```powershell
   .\tools\format.ps1
   .\tools\quality.ps1
   ```

2. Review `git diff`, `git diff --check`, and `git status`. Do not include captures, credentials, build outputs, resource packs, or unrelated files.
3. Update documentation in the same change when behavior, CLI use, capture fields, or a research assumption changed.
4. Add exactly one commit audit record under [`docs/audits/`](docs/audits/) for the authored commit. Include intent, scope, evidence, automated checks, live validation status, limitations, and follow-up work. Use the stable change ID from the audit template rather than trying to predict the commit hash.
5. Commit the coherent change. A runtime validation build must come from a clean committed revision.
6. Decide whether manual validation is required.

   - Documentation, comments, and formatting-only changes may be exempt.
   - Code or any possible runtime behavior change requires The Hive validation.
   - A change to authentication, Experience routing, networking, protocol handling, resource packs, or another specific flow must exercise that flow directly.

7. For a runtime change, build the exact commit and validate it before calling the pull request merge-ready.

   ```powershell
   .\tools\build.ps1 -Version pr
   .\bin\bedrock-debug-proxy.exe version
   ```

   Follow the capture and report steps in [`docs/validation.md`](docs/validation.md). Upload the sanitized report as a pull request artifact or provide it through the review system without adding it to the candidate commit. Keep the raw capture local.
8. Push the branch and open the pull request using the repository template. State what changed, what did not change, automated results, manual results, and any remaining uncertainty.

## After opening a pull request

1. After every push, open the Actions page and inspect the run for the exact commit. Wait for both GitHub Actions jobs to finish. A local quality pass is not a CI result.
2. If a job fails, inspect its logs, classify the failure, make a focused fix with a new audit record, push again, and repeat. Do not declare the change complete while the exact pushed commit is pending or failing.
3. Confirm the pull request template identifies the validation impact.
4. For a runtime change, confirm the The Hive report says `pass` and its `tested_revision` is the exact pull request head commit.
5. Confirm every flow-specific change was observed directly. `not_observed` means the pull request is still awaiting validation.
6. Address review findings with focused commits. Do not hide unrelated refactors in a review fix.
7. If code changes after manual validation, rebuild and repeat the affected live tests. A report for an older commit is not evidence for the new head.
8. Re-run `tools/quality.ps1` after the final change.

A pull request is merge-ready only when scope and provenance are clear, automated checks pass, required live reports match the head revision, review findings are resolved, documentation is current, and no unexplained protocol or decode anomaly remains.

## When a live test fails

Record the result as `fail` or `incomplete` and investigate before changing code. Check the capture, packet definitions, optional fields, packet ordering assumptions, transport, authentication, and whether the server behavior is valid. Do not add a server-specific workaround only to make the gate pass.

The detailed report schema, sensitive-data boundary, and pass criteria are in [`docs/validation.md`](docs/validation.md). Use [`docs/server-targets.md`](docs/server-targets.md) for the active endpoint catalog and server-policy disclaimer. CubeCraft is intentionally excluded from validation and release recommendations because its published policy prohibits Bedrock proxy use. The release process is separate and documented in [`docs/releasing.md`](docs/releasing.md).
