# Test coverage baseline

This page records Biblios' initial automated coverage measurement and the rules
used to improve it. Coverage is evidence about exercised code, not a substitute
for behavioral, integration, security, or API-contract tests.

## How to generate reports

### Backend

```sh
cd backend
go test -race -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

The CI workflow runs the same command and retains `coverage.out` as the
`backend-coverage` artifact for 14 days.

### Frontend

```sh
cd frontend
npm ci
npm run test:coverage
```

Vitest uses the V8 provider and writes text, JSON summary, and LCOV reports to
`frontend/coverage`. CI retains the directory as the `frontend-coverage`
artifact for 14 days.

## Initial baseline

| Area | Measurement | Source | Interpretation |
| --- | ---: | --- | --- |
| Backend | 5.2% statements | Backend CI run on the Sprint 00 baseline | Most production packages and failure paths remain untested. |
| Frontend | 62.14% statements, 54.90% branches, 58.82% functions, 63.04% lines | Vitest V8 run with 52 tests on the Sprint 00 baseline | API books/search and page-level flows remain the largest visible gaps. |

The frontend measurement was generated with the locked dependencies and
`npm run test:coverage` on the Sprint 00 baseline. The first successful
coverage-enabled CI run should confirm it; later changes must update the table
with the source revision and command used.

## Critical-flow gap register

Coverage work is prioritized by impact rather than by percentage:

| Priority | Flow | Required evidence | Owner |
| --- | --- | --- | --- |
| P0 | Authentication, token/session revocation, and authorization boundaries | Unit tests plus handler/integration tests for success, invalid credentials, expired/revoked tokens, and cross-user access | Backend |
| P0 | Book/library/collection writes | Service and repository tests for validation, transactions, duplicate/conflict handling, and rollback | Backend |
| P1 | Reading activity, reviews, shelves, and notifications | Service/handler tests covering empty, malformed, unauthorized, and pagination cases | Backend |
| P1 | API client error and cancellation behavior | Frontend API tests for non-2xx responses, malformed payloads, timeouts, and auth expiry | Frontend |
| P1 | Registration, login, profile, and primary navigation | Component/page tests for validation, loading, failure, and keyboard-accessible states | Frontend |
| P2 | Search, discovery, themes, and secondary UI states | Component tests and focused interaction tests | Frontend |

## Gate and ratchet policy

1. Sprint 00 establishes reproducible reports and the gap register; it does not
   impose an arbitrary percentage gate on legacy code.
2. New or materially changed critical-flow code must include focused tests in
   the same pull request.
3. A numeric gate is introduced after two comparable baseline runs. The first
   gate is the measured baseline rounded down to a conservative whole number.
4. The gate may only increase, and only when the measured default-branch
   coverage is above the proposed value for two consecutive runs.
5. Coverage decreases on changed critical paths block merge even if the global
   percentage remains above the gate.
6. Reports must never include credentials, tokens, production data, or uploaded
   covers. Artifacts are short-lived CI evidence, not a data store.

## Review checklist

- [ ] Backend and frontend reports were generated from the same commit.
- [ ] New critical behavior has a focused test or a documented follow-up issue.
- [ ] Authorization and data-integrity paths were sampled manually.
- [ ] No secret, token, or production fixture appears in generated artifacts.
- [ ] The gap register and Jira follow-up issues are updated when priorities change.
