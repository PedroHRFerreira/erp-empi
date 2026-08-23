.PHONY: setup dev up down test test-go test-frontend lint format backup-verify

setup:
	cp -n .env.example .env || true

dev:
	docker compose up --build

up:
	docker compose up -d --build

down:
	docker compose down

test: test-go test-frontend

test-go:
	go test ./...

test-frontend:
	pnpm --dir frontend test

lint:
	go test ./...
	pnpm --dir frontend lint

format:
	gofmt -w cmd config internal
	pnpm --dir frontend format

backup-verify:
	@test -n "$(BACKUP_PATH)" || (echo "BACKUP_PATH is required" >&2; exit 1)
	./scripts/predeploy-migrate.sh verify "$(BACKUP_PATH)"
