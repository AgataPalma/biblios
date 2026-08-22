# Dependency security audit

Dependency security is checked from the committed lockfiles and module
checksums. A clean result is required before a release candidate can reach
staging.

## Current result

Baseline: 2026-08-22, BLIO-72 remediation branch.

| Supply-chain area | Check | Result | Notes |
| --- | --- | --- | --- |
| Go application | `govulncheck ./...` | 0 reachable vulnerabilities | The scanner reported three vulnerabilities in imported/required packages that are not reachable from application code; they remain transitive inventory items and must be reviewed after future upgrades. |
| npm/frontend | `npm audit --audit-level=high` | 0 vulnerabilities | Direct vulnerable versions of Axios, React Router, PostCSS, Vite, and Vitest/browser tooling were upgraded. |
| Backend image | Clean Go 1.25.13 Alpine build plus tests | Passed | Image rebuild and backend tests completed successfully. |
| Frontend image | Clean npm install during image build | Passed | Image build reported 0 npm vulnerabilities. |

## CI enforcement

- Backend CI runs `go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...`.
- Frontend CI runs `npm audit --audit-level=high` after `npm ci`.
- Container CI rebuilds the complete Compose topology and verifies service
  health before merge.
- Lockfiles (`frontend/package-lock.json`) and Go checksums (`backend/go.sum`)
  are authoritative; dependency updates must include them in the same change.

## Review rules

1. Every advisory is classified as fixed, reachable, not reachable with
   evidence, or mitigated with an owner and review date.
2. Do not use `npm audit fix --force` or broad major-version upgrades without
   running the complete frontend test, build, and container checks.
3. Do not suppress a reachable Go finding. Upgrade, replace, or document a
   time-bounded mitigation with explicit approval.
4. Rebuild and scan runtime images after dependency changes; old image tags are
   not considered remediated merely because source files changed.
5. Generated reports and logs must not contain credentials, tokens, or private
   production data.
