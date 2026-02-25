all: build test

setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
	fi
	@if ! grep -q "^ENCRYPTION_KEY=[A-Za-z0-9+/=]\{40,\}" .env 2>/dev/null; then \
		echo "Generating ENCRYPTION_KEY..."; \
		ENCRYPTION_KEY=$$(openssl rand -base64 32); \
		sed -i.bak "s|^ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$$ENCRYPTION_KEY|" .env && rm -f .env.bak; \
	else \
		echo "ENCRYPTION_KEY already set"; \
	fi
	@if ! grep -q "^HASH_KEY=[A-Za-z0-9+/=]\{40,\}" .env 2>/dev/null; then \
		echo "Generating HASH_KEY..."; \
		HASH_KEY=$$(openssl rand -base64 32); \
		sed -i.bak "s|^HASH_KEY=.*|HASH_KEY=$$HASH_KEY|" .env && rm -f .env.bak; \
	else \
		echo "HASH_KEY already set"; \
	fi
	@echo "Setup complete! Run 'make migrate-up && make run' to start."

build:
	@echo "Building..."
	@go build -o bin/api cmd/api/main.go
	@go build -o bin/migrate cmd/migrate/main.go
	@go build -o bin/qr-worker cmd/qr-worker/main.go
	@go build -o bin/worker cmd/worker/main.go

run:
	@go run cmd/api/main.go

worker:
	@go run cmd/worker/main.go

test:
	@echo "Testing..."
	@go test ./... -v

itest:
	@echo "Running integration tests..."

migrate-up:
	@echo "Running migrations..."
	@go run cmd/migrate/main.go -action=up

migrate-down:
	@echo "Rolling back last migration..."
	@go run cmd/migrate/main.go -action=down -steps=1

migrate-fresh:
	@echo "Running fresh migrations..."
	@go run cmd/migrate/main.go -action=fresh

migrate-status:
	@echo "Checking migration status..."
	@go run cmd/migrate/main.go -action=status

docs:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || (echo "Error: swag is not installed." && echo "Install it with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	@swag fmt
	@swag init -g cmd/api/main.go -o docs

clean:
	@echo "Cleaning..."
	@rm -rf bin

.PHONY: all setup build run run-worker worker test clean itest migrate-up migrate-down migrate-fresh migrate-status docs
