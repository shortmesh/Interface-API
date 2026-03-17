all: build test

setup:
	@./scripts/setup-dev.sh

build: docs
	@./scripts/build.sh

run:
	@./scripts/run-with-tags.sh cmd/api/main.go

worker:
	@./scripts/run-with-tags.sh cmd/worker/main.go

test:
	@echo "Testing..."
	@go test ./... -v

itest:
	@echo "Running integration tests..."

migrate-up:
	@echo "Running migrations..."
	@./scripts/run-with-tags.sh cmd/migrate/main.go -action=up

migrate-down:
	@echo "Rolling back last migration..."
	@./scripts/run-with-tags.sh cmd/migrate/main.go -action=down -steps=1

migrate-status:
	@echo "Checking migration status..."
	@./scripts/run-with-tags.sh cmd/migrate/main.go -action=status

docs:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || (echo "Error: swag is not installed." && echo "Install it with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	@swag fmt
	@swag init -g cmd/api/main.go -o docs

clean:
	@echo "Cleaning..."
	@rm -rf bin

setup-systemd:
	@./scripts/setup-systemd.sh

update-env:
	@./scripts/update-env.sh

update-env-prod:
	@echo "Updating production .env file..."
	@sudo ./scripts/update-env.sh --production

.PHONY: all setup build run worker test clean itest migrate-up migrate-down migrate-status docs setup-systemd update-env update-env-prod
