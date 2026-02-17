all: build test

build:
	@echo "Building..."
	@go build -o bin/api cmd/api/main.go
	@go build -o bin/migrate cmd/migrate/main.go
	@go build -o bin/consumer cmd/consumer/main.go

run:
	@go run cmd/api/main.go

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
	@swag init -g cmd/api/main.go -o docs

clean:
	@echo "Cleaning..."
	@rm -rf bin

.PHONY: all build run test clean itest migrate-up migrate-down migrate-fresh migrate-status docs
