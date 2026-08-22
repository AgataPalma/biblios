# Runtime security and credential operations

## Purpose

This guide defines Biblios runtime modes, local network boundaries, required
secrets, safe bootstrap, rotation, session invalidation, and validation. It is
the source-controlled companion to the authoritative Confluence security and
operations documentation.

No command in this guide requires placing an actual secret in Git, Jira,
Confluence, screenshots, logs, or shell history.

## Security model

Biblios uses three controls together:

1. The backend signs HS256 JSON Web Tokens with `JWT_SECRET`.
2. Every accepted token must also have an active session key in Redis.
3. PostgreSQL stores users, roles, libraries, books, and other durable data.

The controls are not independent if an attacker knows the JWT signing secret
and can write to Redis. They could forge a token and create its matching active
session key. For that reason, the signing secret must be unpredictable, Redis
must require authentication, and both datastores must remain unreachable from
unauthorized networks.

## Environment matrix

| Concern | Development | Test | Staging | Production |
| --- | --- | --- | --- | --- |
| `APP_ENV` | `development` | `test` | `staging` | `production` |
| Secret source | Untracked local `.env` | Isolated test environment | External secret delivery required | External secret delivery required |
| Host-published PostgreSQL | Loopback only when needed | Isolated runner only | Forbidden | Forbidden |
| Host-published Redis | Loopback only when needed | Isolated runner only | Forbidden | Forbidden |
| pgAdmin | Opt-in `tools` profile, loopback only | Disabled | Disabled by default | Disabled |
| Backend/frontend exposure | Loopback | Isolated runner | Ingress design required | TLS ingress required |
| Known/default secrets | Rejected | Rejected | Rejected | Rejected |
| Redis authentication | Required | Required | Required | Required |

The `docker-compose.nonlocal.yml` overlay removes every direct host port. It is
a security assertion used to validate isolation, not a complete staging or
production deployment. It deliberately leaves ingress, TLS, secret delivery,
backups, monitoring, and rollback unspecified.

## Required variables

| Variable | Consumer | Rule |
| --- | --- | --- |
| `APP_ENV` | Backend | Required; one of development, test, staging, production |
| `POSTGRES_DB` | PostgreSQL/backend Compose | Required |
| `POSTGRES_USER` | PostgreSQL/backend Compose | Required |
| `POSTGRES_PASSWORD` | PostgreSQL/backend | Unique; must not be a known default or template value |
| `REDIS_PASSWORD` | Redis/backend | Required; at least 16 characters; not embedded in `REDIS_URL` |
| `JWT_SECRET` | Backend | Required; at least 32 characters; unique per environment |
| `PGADMIN_EMAIL` | Optional pgAdmin | Required only to use the `tools` profile |
| `PGADMIN_PASSWORD` | Optional pgAdmin | Unique and non-default |
| `GOOGLE_BOOKS_API_KEY` | Backend lookup | Optional |

`DATABASE_URL` and `REDIS_URL` are assembled by local Compose. Direct backend
execution must provide them explicitly. `REDIS_URL` contains only location and
transport; Redis credentials are supplied separately through
`REDIS_PASSWORD` so configuration errors cannot create two competing password
sources.

## New local bootstrap

1. Create the untracked environment file:

   ```bash
   cp .env.example .env
   chmod 600 .env
   ```

2. Generate a different random value for each password and secret. One safe
   generator is:

   ```bash
   openssl rand -hex 32
   ```

   Run it separately for PostgreSQL, Redis, JWT signing, and pgAdmin. Copy the
   results directly into `.env`. Do not paste them into tickets or documents.

3. Confirm `.env` is ignored:

   ```bash
   git status --short --ignored .env
   ```

4. Validate interpolation without printing the rendered configuration:

   ```bash
   docker compose config --quiet
   ```

5. Start the core services:

   ```bash
   docker compose up --build
   ```

6. Verify service state and host bindings:

   ```bash
   docker compose ps
   ```

   Published addresses must start with `127.0.0.1:`. PostgreSQL and Redis must
   not show `0.0.0.0` or `[::]` host bindings.

## Rotating an existing local environment

Changing `POSTGRES_PASSWORD` in `.env` does not change the password inside an
existing PostgreSQL data volume. Rotate the database first, then update the
environment file.

1. Keep the existing stack running and open PostgreSQL's interactive client:

   ```bash
   docker compose exec postgres sh -c 'psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB"'
   ```

2. At the `psql` prompt, run `\password` and enter the new randomly generated
   database password twice. The prompt does not echo the value.
3. Exit with `\q` and update `POSTGRES_PASSWORD` in `.env`.
4. Replace `REDIS_PASSWORD`, `JWT_SECRET`, and `PGADMIN_PASSWORD` in `.env`
   with separately generated values.
5. Recreate the services so they receive the new configuration:

   ```bash
   docker compose up -d --build --force-recreate postgres redis backend frontend
   ```

6. Invalidate the Redis session allow-list after Redis is healthy:

   ```bash
   docker compose exec redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli FLUSHDB'
   ```

   Redis is currently dedicated to session state. If another data class is
   added later, replace `FLUSHDB` with a scoped session-key deletion procedure.

7. Recreate optional pgAdmin when it is next needed:

   ```bash
   docker compose --profile tools up -d --force-recreate pgadmin
   ```

8. Log in again. Tokens signed with the previous JWT secret must fail, and the
   cleared session allow-list must contain no old sessions.

For a disposable local database, a clean bootstrap may be simpler. Removing
`postgres_data` destroys local application data and therefore requires an
explicit backup/recovery decision; never use `docker compose down -v` casually.

## Negative configuration validation

The backend validates configuration before connecting to PostgreSQL or Redis.
Automated tests cover missing variables, unknown runtime modes, short or
default JWT secrets, default database passwords, Redis credentials embedded in
URLs, missing Redis authentication, and example placeholders.

Run them with:

```bash
cd backend
go test ./internal/config
```

Validate the non-local port boundary without printing secrets:

```bash
APP_ENV=staging docker compose \
  -f docker-compose.yml \
  -f docker-compose.nonlocal.yml \
  config --quiet
```

To inspect only published-port declarations, use a reviewed redaction-safe
query. Do not paste full `docker compose config`, `docker inspect`, or container
environment output into shared systems because rendered values can include
credentials.

## Secret scanning

The `Secret scan` GitHub Actions workflow scans complete Git history on every
pull request, every push to `main`, and manual dispatch. It uses the official
Gitleaks container with redaction enabled. Both the checkout action and scanner
container are pinned to immutable identifiers; upgrades must be reviewed and
tested rather than following mutable tags automatically.

Run the equivalent scan locally when Docker can pull the pinned image:

```bash
docker run --rm \
  --volume "$PWD:/repo" \
  --workdir /repo \
  ghcr.io/gitleaks/gitleaks:v8.29.1@sha256:aa036a2f4bdfe3cc3c55fa4326308efabb4a6be498c883c864fd1d0d5585438a \
  git --redact --no-banner /repo
```

A passing scanner reduces accidental disclosure risk but does not prove that
the history has never contained a secret. A discovered credential must be
rotated even if its commit is later removed or rewritten.

## Redis authentication checks

An unauthenticated command must fail:

```bash
docker compose exec redis redis-cli ping
```

An authenticated health check using the container's existing environment must
succeed without putting the value in host shell history:

```bash
docker compose exec redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli ping'
```

Expected results are an authentication error for the first command and `PONG`
for the second. Record only those outcomes.

## Non-local deployment gate

Do not deploy Biblios outside an isolated developer machine until all of the
following exist and are verified:

* staging or production `APP_ENV`;
* secrets delivered outside Git and Compose files;
* no direct PostgreSQL, Redis, or pgAdmin host publication;
* authenticated Redis on a private network;
* TLS ingress and reviewed CORS/origin configuration;
* firewall/security-group or equivalent network policy;
* credential rotation and session invalidation procedure;
* backups, restore testing, monitoring, alerting, and rollback;
* a penetration-style regression proving a forged JWT cannot be paired with an
  attacker-created session key.

The non-local overlay alone does not satisfy this gate.

## Incident and evidence rules

If a credential may have been disclosed, rotate it; do not merely delete the
message or log containing it. Rotate the JWT secret and clear sessions together.
Assess database and Redis access logs where available. Document dates, variable
names, affected environments, actions, and validation outcomes, but never the
old or new values.
