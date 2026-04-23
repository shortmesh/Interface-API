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

format:
	@echo "Formatting codebase..."
	@echo ""
	@echo "Running go fmt..."
	@go fmt ./... || echo "WARNING: go fmt encountered issues (continuing...)"
	@echo ""
	@echo "Running prettier on HTML, CSS, JS..."
	@if command -v npx >/dev/null 2>&1; then \
		echo "Formatting HTML files..."; \
		find ./pkg/adminweb/web -name "*.html" -type f -exec npx prettier --write {} \; 2>/dev/null || echo "WARNING: prettier HTML formatting encountered issues (continuing...)"; \
		echo "Formatting CSS files..."; \
		find ./pkg/adminweb/web -name "*.css" -type f -exec npx prettier --write {} \; 2>/dev/null || echo "WARNING: prettier CSS formatting encountered issues (continuing...)"; \
		echo "Formatting JS files..."; \
		find ./pkg/adminweb/web -name "*.js" -type f -exec npx prettier --write {} \; 2>/dev/null || echo "WARNING: prettier JS formatting encountered issues (continuing...)"; \
	else \
		echo "WARNING: npx not found. Install Node.js to use prettier"; \
	fi
	@echo ""
	@echo "Running swag fmt..."
	@which swag > /dev/null && swag fmt 2>/dev/null || echo "WARNING: swag fmt encountered issues (continuing...)"
	@echo ""
	@echo "Formatting complete!"

clean:
	@echo "Cleaning..."
	@rm -rf bin

setup-systemd:
	@./scripts/setup-systemd.sh

.PHONY: all setup build run worker test clean itest migrate-up migrate-down migrate-status docs format setup-systemd
