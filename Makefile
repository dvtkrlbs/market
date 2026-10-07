.PHONY: db-up db-down generate hooks migrate-apply openapi run test

db-up:
	docker compose up -d --wait postgres

db-down:
	docker compose down

generate:
	go generate ./...

hooks:
	go tool lefthook install

migrate-apply:
	go run ./cmd/market migrate

openapi:
	go run ./cmd/openapi-stitch

run:
	go run ./cmd/market serve

test:
	go test ./...
