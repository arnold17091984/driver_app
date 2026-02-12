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

migrate-up:
	cd backend && go run cmd/migrate/main.go up

migrate-down:
	cd backend && go run cmd/migrate/main.go down

seed:
	cd backend && go run cmd/migrate/main.go seed

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
