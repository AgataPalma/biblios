# Biblios API — Postman Test Collection

## Purpose

`biblios-api.postman_collection.json` is the source-controlled executable HTTP collection for the Biblios backend. It serves four related purposes:

1. Manual endpoint exploration during development.
2. Executable request and response examples.
3. Regression checks at the deployed HTTP boundary.
4. Coverage input for automated router, OpenAPI, and collection drift detection.

The collection is not the only API test layer. Go unit and integration tests remain responsible for internal business rules, transactions, concurrency, repository failures, and security invariants that cannot be proven reliably by an external HTTP collection alone.

## Current verified baseline

The automated validator currently reports:

| Surface | Current state |
| --- | ---: |
| Registered HTTP operations | 89 |
| Postman requests | 107 |
| Registered operations represented by Postman | 89 of 89 |
| Requests with an active post-response script | 107 of 107 |
| Requests with an explicit status assertion | 107 of 107 |
| OpenAPI operations | 28 |
| Known OpenAPI gaps | 61 |

Endpoint representation means that a request exists for the method and path. It does not prove that every request has complete schema, authorization, privacy, lifecycle, and business-semantic assertions. BLIO-47 owns that deeper coverage classification.

The OpenAPI gap is recorded explicitly in `api-tests/postman/openapi-gap-baseline.json`. The baseline is a temporary ratchet, not an acceptable final contract. CI prevents it from changing silently, and BLIO-43 must reduce it to zero.

## Authoritative files

| File | Responsibility |
| --- | --- |
| `backend/cmd/api/main.go` | Implementation route inventory |
| `backend/cmd/api/docs/openapi.yaml` | Reviewed public API contract |
| `biblios-api.postman_collection.json` | Executable requests and post-response tests |
| `api-tests/postman/environments/local.postman_environment.json` | Non-production local template |
| `api-tests/postman/environments/ci.postman_environment.json` | Empty CI template; runtime values must be injected |
| `api-tests/postman/openapi-gap-baseline.json` | Reviewed, temporary list of missing OpenAPI operations |
| `scripts/validate-api-contract.mjs` | Drift, safety, variable, and minimum-test validation |
| `.github/workflows/api-contract.yml` | Pull-request and main-branch automation |

Do not create another independent collection. Update the source-controlled collection and import it into Postman when needed.

## Safety rules

- Run state-changing requests only against disposable local, CI, or dedicated test databases.
- Never point the collection at production.
- Never commit real passwords, JWTs, API keys, invitation tokens, private identifiers, or production URLs.
- Collection token and identifier variables must remain empty in Git.
- CI credentials must be short-lived or repository secrets and must be injected at runtime.
- Do not export a personal Postman environment over the committed templates.
- Review collection diffs like application code.
- A passing status assertion alone is not proof that the endpoint behaved correctly.

## Local prerequisites

- Docker and Docker Compose
- Biblios backend, PostgreSQL, and Redis running
- Postman desktop application or Postman CLI/Newman
- A disposable local database
- A non-production moderator account for privileged folders

The default local addresses are:

| Service | URL |
| --- | --- |
| API base | `http://localhost:8081/api/v1` |
| Health endpoint | `http://localhost:8081/health` |

## Import into Postman

1. Start the required Biblios containers.
2. Open Postman.
3. Select **Import**.
4. Import `biblios-api.postman_collection.json`.
5. Import `api-tests/postman/environments/local.postman_environment.json`.
6. Select the **Biblios — Local** environment.
7. Review all environment values before sending a request.
8. Configure moderator credentials only in your local Postman current values. Do not export them back to Git.
9. Run the Health folder first.

## Environment variables

### Configuration and test identities

| Variable | Meaning | Committed value policy |
| --- | --- | --- |
| `baseUrl` | Base URL including `/api/v1` | Local or CI URL allowed |
| `healthUrl` | Root health URL without `/api/v1` | Local or CI URL allowed |
| `testUserEmail` | Standard test identity email | Synthetic non-production value only |
| `testUsername` | Standard test identity username | Synthetic value only |
| `testUserPassword` | Standard test identity password | Local dummy or runtime-injected value only |
| `updatedUserEmail` | Email used by update-email scenarios | Synthetic value only |
| `updatedUserPassword` | Password used by update-password scenarios | Local dummy or runtime-injected value only |
| `duplicateUsername` | Username used by duplicate-email tests | Synthetic value only |
| `invalidPassword` | Intentionally wrong password | Local dummy value only |
| `moderatorEmail` | Pre-provisioned moderator identity | Empty in Git |
| `moderatorPassword` | Pre-provisioned moderator password | Empty in Git |
| `inviteeEmail` | Invitation target | Synthetic non-production value only |

### Runtime values

The collection stores these values only in its in-memory collection-variable scope. Their committed values must remain empty:

- `authToken`
- `modToken`
- `userId`
- `bookId`
- `editionId`
- `copyId`
- `libraryId`
- `collectionId`
- `shelfId`
- `reviewId`
- `submissionId`
- `challengeId`
- `sessionId`
- `contributorId`
- `seriesId`
- `invitationToken`
- `notificationId`
- `targetUserId`

## Collection organization

| Folder | Responsibility |
| --- | --- |
| Auth | Registration, login, current user, and logout |
| Users | Profile, email, password, theme, and account deletion |
| Books | Catalogue queries, lookup, submissions, editions, and copies |
| Libraries | Libraries, books, members, permissions, and invitations |
| Invitations | Invitation listing and lifecycle transitions |
| Collections | Cooperative collection lifecycle and membership |
| Reviews | Review lifecycle and likes |
| Notifications | Listing and read state |
| Reading | Challenges, sessions, progress, and statistics |
| Shelves | Shelf lifecycle and book membership |
| Contributors | Contributor discovery and creation |
| Series | Series discovery and creation |
| Genres & Moods | Taxonomy discovery and privileged creation |
| Moderation | Submission decisions and audit logs |
| Admin | Privileged catalogue maintenance |
| Health | Root service health |
| Error Cases | Selected authentication, validation, authorization, and not-found failures |

## Running tests manually

### Safe first check

Run the Health folder. It is read-only and does not require authentication.

### Standard authenticated flow

For a fresh disposable database:

1. Run **Register User**.
2. Confirm the response contains a token and user identifier.
3. Run **Get Current User**.
4. Execute only the domain scenarios whose prerequisites are satisfied.
5. Run cleanup requests last.

### Privileged flow

Moderator and admin scenarios require a deliberately provisioned non-production account. Configure `moderatorEmail` and `moderatorPassword` in local current values, then run **Login as Moderator** before privileged requests.

### Important ordering limitation

The collection contains destructive and stateful requests. The entire collection is not yet a deterministic, one-click regression run because several folders share resources and some requests delete those resources. Full deterministic execution requires fixture orchestration, unique run identities, privileged-user provisioning, dependency ordering, and guaranteed cleanup. Until that work is complete:

- Run folders or documented workflows deliberately.
- Do not interpret an arbitrary full collection run as release evidence.
- CI runs structural coverage checks for every request and an executable Health smoke test.
- BLIO-47 and the backend/container CI items own the remaining deterministic runtime coverage work.

## Command-line execution

### Contract and collection validation

Run:

```bash
node scripts/validate-api-contract.mjs
```

The command fails when:

- A Go route has no Postman request.
- A Postman request does not match a registered route.
- An OpenAPI gap changes without updating the reviewed baseline.
- OpenAPI contains an operation not registered by the router.
- A request has no active test script.
- A request has no explicit status assertion.
- A request contains a literal authorization or API-key value.
- A referenced Postman variable is undeclared.
- A sensitive collection variable contains a committed value.

To prove whether OpenAPI is fully reconciled, run:

```bash
node scripts/validate-api-contract.mjs --strict-openapi
```

This strict command is expected to fail until the 61 known OpenAPI gaps are resolved.

### Newman

The CI workflow uses the pinned official Newman package. A local Health run is:

```bash
npx --yes newman@6.2.2 run biblios-api.postman_collection.json \
  --environment api-tests/postman/environments/local.postman_environment.json \
  --folder "🏥 Health" \
  --bail
```

Do not run the full state-changing suite against a persistent environment until deterministic fixture orchestration is implemented.

## Continuous integration

`.github/workflows/api-contract.yml` runs on every pull request, every push to `main`, and manual dispatch.

### Contract coverage job

This job runs without containers and validates:

- Go route inventory
- OpenAPI inventory and ratcheted gaps
- Postman method/path coverage
- Minimum request-test requirements
- Variable declarations
- Credential safety rules

### Runtime smoke job

This job:

1. Starts isolated PostgreSQL, Redis, and backend containers.
2. Waits for `/health` to become available.
3. Runs the source-controlled Health folder through Newman 6.2.2.
4. Produces CLI, JUnit, and JSON results.
5. Prints service logs on failure.
6. Removes the isolated containers and volumes even when a step fails.

The contract job enforces collection updates for all endpoint changes. The runtime job currently proves that the committed collection can execute against a composed Biblios backend. Broader runtime execution must be added incrementally as deterministic fixtures become available.

## Endpoint-change checklist

Any branch that changes an endpoint must update all affected artifacts in the same pull request:

1. Go route registration and handler behavior.
2. Request and response DTOs.
3. OpenAPI path, method, parameters, security, request body, responses, and examples.
4. Postman request URL, method, headers, body, description, and tests.
5. Positive, validation, authentication, authorization, ownership, conflict, and lifecycle scenarios as applicable.
6. Frontend client and TypeScript types when consumed by the frontend.
7. Domain API documentation in Confluence.
8. Jira acceptance criteria and validation evidence.

Run the validator before committing. CI rejects unreviewed drift.

## Adding a new endpoint

1. Assign a stable OpenAPI `operationId`.
2. Place the request in the correct domain folder.
3. Use `baseUrl` or `healthUrl`; do not hard-code a hostname.
4. Use variables for identifiers and test identities.
5. Add an unconditional status assertion.
6. Add response-schema or semantic assertions appropriate to the endpoint.
7. Capture generated identifiers only after verifying the successful response.
8. Add applicable negative and authorization scenarios.
9. Document prerequisites and cleanup.
10. Run `node scripts/validate-api-contract.mjs`.

## Removing or renaming an endpoint

1. Update the router and implementation.
2. Remove or deprecate the OpenAPI operation.
3. Remove or update every matching Postman request, including negative scenarios.
4. Search frontend clients and documentation for the old path.
5. Run the validator.
6. Record compatibility and migration impact in Jira and Confluence.

## Definition of complete API test coverage

An operation is considered structurally covered when its method and path are represented in the collection and it has an active status assertion.

An operation is considered behaviorally covered only when all applicable categories have evidence:

- Success response
- Request validation
- Authentication
- Role authorization
- Ownership or cooperative-library permissions
- Not found
- Conflict or duplicate behavior
- Pagination, filtering, and sorting
- State transitions and idempotency
- Response schema and standard errors
- Sensitive-data handling
- Cleanup and repeatability

The long-term target is 100% structural coverage and explicit applicable/not-applicable decisions for every behavioral category.

## Known gaps

- OpenAPI describes only 28 of 89 registered operations.
- Sixty-one requests currently have status-only tests and require classification under BLIO-47.
- Full-suite fixture orchestration is not deterministic.
- Moderator and admin provisioning is manual outside the Health CI smoke path.
- The cover-upload request requires an explicitly selected local file and is not part of automated execution.
- Automated Confluence publication of run summaries is not yet connected to CI credentials.

These gaps must remain visible in Jira and Confluence until resolved; they must not be described as completed coverage.
