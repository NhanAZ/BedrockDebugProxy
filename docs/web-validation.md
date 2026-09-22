# Web viewer validation

The capture viewer is a user-facing research surface. A broken HTML, CSS, JavaScript asset, API response, or interaction can make packet evidence unavailable even when the proxy and capture remain healthy.

## Same-change rule

Any change that may affect the viewer must update the page and its validation in the same coherent change. This includes files under `internal/viewer`, embedded assets, the `view` command, `run --web-viewer`, viewer HTTP handlers, and capture fields consumed by the page. Do not defer the viewer repair to a later commit or treat a successful Go build as evidence that the page works.

## Automated checks

Run the project gate from a clean checkout or before committing:

```powershell
.\tools\format.ps1
.\tools\quality.ps1
```

The gate performs all of the following for the embedded viewer:

- serves and checks the HTML, CSS, JavaScript, and font routes through the viewer HTTP test;
- verifies that the page links its required assets and filter controls;
- checks the embedded JavaScript with `node --check internal/viewer/assets/app.js`;
- runs the complete Go test, build, static-analysis, vulnerability, license, artifact, and whitespace checks.

The quality workflow runs the same gate on GitHub Actions. A local pass is not a substitute for the required post-push CI result.

## Browser smoke check

For a presentation or interaction change, commit the change and build the exact revision:

```powershell
.\tools\build.ps1 -Version web-check
.\bin\bedrock-debug-proxy.exe view --no-open CAPTURE_DIRECTORY
```

Open the printed loopback URL in a browser. Confirm that the page loads without a blank or partially rendered state and that the browser console has no errors. Exercise the affected controls. For the general viewer, also select a filter, reset it, open an event, inspect decoded details, inspect a blob preview, and verify that the page remains read-only. For `view --live` or `run --web-viewer`, confirm that new complete events appear, Follow live selects the newest matching event and scrolls it into view, and the final verification state is reached after the capture closes.

Record the exact binary revision, capture identifier, browser smoke result, and any remaining platform-specific rendering limits in the commit audit. Keep captures and sensitive viewer URLs out of Git.
