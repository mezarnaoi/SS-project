# Security: Patch Vite high-severity advisory (GHSA-p9ff-h696-f583)

## Summary

`client` used `vite@6.3.5`, which is affected by GHSA-p9ff-h696-f583 (high severity: arbitrary file read via Vite dev server WebSocket).

## Impact

- Risk category: Information disclosure
- Severity: High
- Scope: Development server (`vite`), relevant if exposed to untrusted clients/networks

## Root cause

Direct dependency in `client/package.json` pinned to a vulnerable Vite release line.

## Proposed fix

- Upgrade `vite` from `6.3.5` to `6.4.2` in `client/package.json`
- Regenerate `client/package-lock.json`
- Add documentation in `docs/VULNERABILITY_VITE_HIGH_FIX.md`

## Validation

- `npm audit` before: 7 high vulnerabilities
- `npm audit` after: 6 high vulnerabilities
- `vite` entry removed from audit report
- `npm run build` passes

## Acceptance criteria

- [ ] Vite upgraded to non-vulnerable version
- [ ] Audit no longer reports `vite` vulnerability
- [ ] Frontend build succeeds
- [ ] PR reviewed and merged
