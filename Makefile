ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: dev build run schema seed reset dev-setup test vet tidy

dev:
	air

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

# Apply full schema (creates all tables)
schema:
	psql "$(DATABASE_URL)" -f db/schema.sql

# Insert initial data (admin user + sample content)
seed:
	psql "$(DATABASE_URL)" -f db/seed.sql

# Drop all tables and re-apply schema + seed (local dev reset)
reset:
	psql "$(DATABASE_URL)" -f db/reset.sql
	psql "$(DATABASE_URL)" -f db/schema.sql
	psql "$(DATABASE_URL)" -f db/seed.sql

# One-command local setup: apply schema then seed
dev-setup: schema seed

vet:
	go vet ./...

test:
	go test ./... -race

tidy:
	go mod tidy
