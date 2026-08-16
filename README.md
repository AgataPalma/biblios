# Biblios

Biblios is a full-stack library and reading application. The Go API manages
accounts, authentication, catalogues, personal and cooperative libraries,
collections, shelves, reading activity, reviews, notifications, moderation,
contributors, genres, moods, and series. The React application provides the
browser interface. PostgreSQL stores durable application data and Redis stores
the active-session allow-list.

## Local development

Local Docker development is intentionally isolated to the host loopback
interface. PostgreSQL, Redis, pgAdmin, the API, and Vite must not be reachable
through a LAN, VPN, or public interface merely because Compose is running.

1. Copy `.env.example` to `.env`.
2. Replace every placeholder with a unique value. Do not reuse passwords.
3. Start the core stack with `docker compose up --build`.
4. Start optional pgAdmin only when needed with
   `docker compose --profile tools up -d pgadmin`.
5. Open the frontend at `http://localhost:5173`.

The application will not start with missing security-sensitive configuration,
known defaults, example placeholders, a JWT secret shorter than 32 characters,
or a Redis password shorter than 16 characters.

See [runtime security and credential operations](docs/security/runtime-configuration.md)
before rotating an existing environment or evaluating a non-local deployment.

## Frontend runtime

Frontend development and the production-style static image are documented in
[the frontend guide](frontend/README.md). The production-style frontend image
does not make the backend or datastores production-ready.
