.PHONY: help install build clean dev preview

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "Installing dependencies..."
	yarn install

build: ## Build the Vite app for production
	@echo "Building Vite app..."
	yarn build
	@echo "Build complete. Output in dist/"

dev: ## Start development server
	@echo "Starting development server..."
	yarn dev

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf dist
	rm -rf node_modules
	@echo "Clean complete"

preview: ## Preview production build locally
	@echo "Starting preview server..."
	yarn preview
