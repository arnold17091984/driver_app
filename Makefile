.PHONY: dev run build migrate-up migrate-down seed db-up db-down test

# Backend
dev:
	cd backend && go run cmd/server/main.go

build:
	cd backend && go build -o bin/server cmd/server/main.go

run: build
	./backend/bin/server

test:
	cd backend && go test ./...

# Database
db-up:
	docker compose up -d postgres

db-down:
	docker compose down

# NOTE: Migrations and seeds run automatically on server startup.
# These targets are kept for documentation but just start the server.
migrate-up:
	@echo "Migrations run automatically on server start (make dev)"

seed:
	@echo "Seed data is applied as migration 000002 on server start (make dev)"

# Web
web-dev:
	cd web && pnpm dev

web-build:
	cd web && pnpm build

# Mobile
mobile-ios:
	cd mobile && npx react-native run-ios

mobile-android:
	cd mobile && npx react-native run-android
