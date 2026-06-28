.PHONY: run build test test-integration lint generate migrate migrate-down migrate-create

DB_URL ?= postgres://polina:polina@localhost:5432/polina?sslmode=disable

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

generate:
	swag init -g cmd/server/main.go -o docs --parseInternal --parseDependency --parseDepth 2

test:
	go test ./... -race -count=1

test-integration:
	go test ./... -tags=integration -race -count=1

lint:
	golangci-lint run ./...

migrate:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name
