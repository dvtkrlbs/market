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

Lefthook is pinned as a Go tool. The pre-commit hook runs `gofmt` on staged Go
files and re-stages them. The pre-push hook checks all tracked Go files without
modifying them.

## Start locally

```sh
make db-up
make migrate-apply
go run ./cmd/market add --description "Opening balance" --amount 100.00
go run ./cmd/market add --description "Coffee" --amount -4.50
go run ./cmd/market list
go run ./cmd/market serve
```

The API listens on `http://localhost:8080` and implements the contract in
`api/openapi.yaml`:

```sh
curl http://localhost:8080/health
curl http://localhost:8080/entries
```

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

- `api/` — OpenAPI contract and oapi-codegen configuration
- `cmd/market/` — API server and bookkeeping CLI entrypoint
- `internal/ent/` — Ent schema and generated database code
- `internal/` — API, domain service, and database wiring
- `migrations/` — embedded golang-migrate SQL migrations
- `web/` — placeholder for future API clients
