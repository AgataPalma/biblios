# Container runtime and local topology

This document describes the supported Compose runtime for Biblios development and the controls that must be preserved when the application is promoted to staging or production.

## Service boundaries

The API and frontend communicate over the private `app` network. PostgreSQL, Redis, and pgAdmin communicate over the private `data` network. Both networks are marked internal; only the API and frontend are application-facing services. Database and Redis ports are bound to loopback for local administration and are not exposed to the LAN.

The development API image runs Air as UID/GID `10001`, with its temporary build directory mounted from a dedicated volume initialized with matching ownership. The production API target is a static binary in a minimal Alpine image and also runs as UID/GID `10001`. Frontend development runs as the unprivileged Node user; the production image runs as the unprivileged nginx user.

## Images and build context

Compose infrastructure images are pinned by digest. Backend and frontend build contexts have dedicated `.dockerignore` files that exclude Git metadata, environment files, dependency directories, generated output, logs, and local test artifacts. Production images should be built from the `production` target and promoted by immutable image digest.

The repository currently validates Go and npm dependency vulnerabilities in CI. Image SBOM generation, image vulnerability scanning, signing, and registry promotion remain deployment work for BLIO-48/BLIO-70 and must be completed before production use.

## Persistence and recovery

Named volumes hold PostgreSQL data, Redis append-only data, and downloaded cover files. The one-shot `redis-data-init` service fixes the Redis volume ownership before Redis starts; it is intentionally separate from the long-running Redis process. Volume names are project-scoped, so do not remove volumes during routine restarts.

Backups and restore drills are not automated by this Compose file. Staging and production must define encrypted PostgreSQL backups, retention, restore verification, cover-file backup strategy, and an explicit Redis recovery policy (Redis is cache/session state unless the product decision says otherwise).

## Startup and migrations

`config-check` validates required runtime configuration before stateful services start. The API currently applies database migrations during API startup. Until a dedicated migration-owner job is introduced, do not scale API replicas concurrently during a migration window. This is tracked as deployment work in BLIO-48.

## Host prerequisites and operations

Redis may warn when Linux memory overcommit is disabled. The host operator should verify `vm.overcommit_memory=1` (for example, with `sysctl vm.overcommit_memory=1`) and persist that setting through the host provisioning system. This is a host-level prerequisite, not something the application container should change.

All long-running services use restart policies, health checks where supported, bounded JSON logs, dropped Linux capabilities, and `no-new-privileges`. These controls reduce blast radius but do not replace network policy, secret management, backups, monitoring, or image admission controls in staging/production.

## Validation commands

Validate interpolation without starting services:

```sh
docker compose --profile tools config --quiet
```

Build the production API image and inspect its runtime identity:

```sh
docker build --target production -t biblios-backend:local-production backend
docker run --rm --entrypoint id biblios-backend:local-production
```

The isolated topology validation should confirm API, frontend, pgAdmin, PostgreSQL, and authenticated Redis health before the project-scoped stack is removed.
