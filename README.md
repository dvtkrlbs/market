# Market

A small Go bookkeeping app with an OpenAPI-first HTTP API, Ent, PostgreSQL,
and embedded migrations powered by golang-migrate. The same binary also
provides bookkeeping CLI commands.

## Prerequisites

- Go 1.27+
- Docker with Compose

The local database is PostgreSQL 18. Ent and oapi-codegen use Go `tool`
directives pinned in `go.mod`.

Enable the repository's Lefthook-managed Git hooks once after cloning:

```sh
make hooks
```

Lefthook is pinned as a Go tool. The pre-commit hook runs `go generate ./...`,
stages its generated API, Ent, and OpenAPI outputs, then runs `gofmt` on staged
Go files and re-stages them. The pre-push hook checks all tracked Go files
without modifying them.

## Start locally

The backend's `.env.example` lists all settings. Copy it to `.env` if needed and
fill `SESSION_SIGNING_KEY` with the output of `openssl rand -base64 32`.
`.env` is ignored by Git; keep the signing key stable across restarts.
`ENVIRONMENT` defaults to `DEVELOPMENT`. In that mode the backend automatically
loads `.env` from the current working directory before reading configuration;
existing process environment variables take precedence. A missing `.env` is
allowed, while an unreadable or malformed file fails startup. Run from the
repository root to use its `.env`. Export `ENVIRONMENT=PRODUCTION` (or another
non-development value) to disable dotenv loading and provide settings through
the process environment instead.

```sh
make db-up
make migrate-apply
go run ./cmd/market add --description "Opening balance" --amount 100.00
go run ./cmd/market add --description "Coffee" --amount -4.50
go run ./cmd/market list
go run ./cmd/market serve
```

The API listens on `http://localhost:8080` and implements the modular contracts
under `api/`.

The server logs each request's method, path, status, duration, and response size
using `slog`, including CORS and validation failures. Handler errors are logged
server-side while clients receive a generic error. Request bodies, headers, and
query strings are not included in request logs.

Example requests:

```sh
curl http://localhost:8080/health
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"person@example.com","password":"correct horse battery staple"}'
curl -c cookies.txt -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"person@example.com","password":"correct horse battery staple"}'
curl -b cookies.txt http://localhost:8080/auth/whoami
curl -b cookies.txt http://localhost:8080/entries
curl -b cookies.txt http://localhost:8080/orders
curl -b cookies.txt 'http://localhost:8080/users/1/stores?for=shop'
```

Login sets an HttpOnly, SameSite=Lax cookie named `session` containing an opaque
database-backed token signed with HMAC-SHA256. Login and refresh return only
`expiresAt` in JSON; the token is sent exclusively in the cookie.
Set `SESSION_SIGNING_KEY` to a base64-encoded random key of at least 32
bytes; keep it secret and reuse it across restarts and server instances. Changing
the key invalidates existing cookies. Unsigned or modified cookies are rejected.
The cookie is Secure on HTTPS
requests. Sessions expire after 30 days; `POST /auth/refresh` extends the same
session and cookie for another 30 days, and `POST /auth/logout` revokes it and
clears the cookie. Protected routes do not accept Authorization headers.
For direct browser API calls, set `CORS_ALLOWED_ORIGINS` to a comma-separated
list of allowed frontend origins (for example `http://localhost:5176`). The
server permits credentialed CORS and unauthenticated preflights only for these
origins; cross-origin requests are denied by default. Do not use wildcards.
Session IP metadata prefers `Forwarded` and `X-Forwarded-For`, so deployments
must configure their reverse proxy to replace untrusted forwarding headers.

## Stores

Stores are global entities with `id`, `name`, and `approved`. All store routes
require a session, and `{userId}` must match the authenticated user:

- `GET /users/{userId}/stores?for=shop` (default): all approved stores.
- `GET /users/{userId}/stores?for=manage`: stores where the user has a
  `StoreAdmin` membership, including unapproved stores but excluding deleted stores.
- `GET /users/{userId}/stores?for=admin`: stores where the user has a
  `StoreAdmin` membership, including unapproved and soft-deleted stores.
- `POST /users/{userId}/stores` with `{"name":"My shop"}`: creates a store
  with `approved=false` and grants the creator a `StoreAdmin` membership.
- `GET /users/{userId}/stores/{storeId}`: an approved store, or any store the
  user administers.
- `PATCH /users/{userId}/stores/{storeId}` with `name` and/or `approved`:
  updates a store the user administers. Changing `approved` also requires
  membership in the global `admins` table.
- `POST /stores/{storeId}/approve`: approves a non-deleted store. The authenticated
  user must be listed in `admins`; a StoreAdmin membership alone is insufficient.
  Repeated approval returns the approved store without additional side effects.
- `DELETE /users/{userId}/stores/{storeId}`: soft-deletes a store the user
  administers by setting `deleted_at`, retaining its `StoreAdmin` memberships.
  Deleted stores are excluded from `shop` and `manage` listings, reads, and updates.

Store creation is open to any authenticated user. StoreAdmin memberships are
unique per user/store pair; stores do not have an owning user.

Apply the admin migration with `make migrate-apply`. Global admins are provisioned
directly in the database for now (no admin membership is granted automatically):

```sql
INSERT INTO admins (user_id) VALUES (1) ON CONFLICT (user_id) DO NOTHING;
```

The approval handler checks current membership on every request. Removing an
`admins` row revokes approval permission without requiring the user to sign out.

The OpenAPI modules are stitched into a validated, self-contained contract with:

```sh
make openapi
```

The command writes `dist/openapi.yaml` and rejects duplicate paths, operation IDs,
and component names across modules.

## Schema workflow

Edit schemas in `internal/ent/schema`, regenerate Ent code, and add matching up/down
migration files under `migrations/`:

```sh
make generate
make migrate-apply
```

Migration files follow the golang-migrate naming convention, for example
`20261007000000_init.up.sql` and `20261007000000_init.down.sql`. They are
embedded into the application binary and applied with `market migrate`.

## Layout

- `api/` — OpenAPI contracts and per-module oapi-codegen configuration
- `cmd/market/` — API server and bookkeeping CLI entrypoint
- `internal/ent/` — Ent schema and generated database code
- `internal/` — API, domain service, and database wiring
- `migrations/` — embedded golang-migrate SQL migrations
- `web/` — Vite+ workspace with the `shared` theme and Orval API client package (see
  [web/README.md](web/README.md))
