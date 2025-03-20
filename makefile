.PHONY: setup dev test migrate build docker-up docker-down

# Configuration
BINARY_NAME=k8s-platform
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=k8s_platform

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@chmod +x ./local-dev-setup.sh
	@./local-dev-setup.sh

# Run development server with hot reload
dev:
	@echo "Starting development server..."
	@air -c .air.toml

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run short tests (skip integration tests)
test-short:
	@echo "Running short tests..."
	@go test -short -v ./...

# Database migrations
migrate-up:
	@echo "Running migrations..."
	@chmod +x ./scripts/migrate.sh
	@./scripts/migrate.sh up

migrate-down:
	@echo "Rolling back migrations..."
	@chmod +x ./scripts/migrate.sh
	@./scripts/migrate.sh down

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	@echo "Creating migration: $(name)"
	@chmod +x ./scripts/migrate.sh
	@./scripts/migrate.sh create $(name)

# Build the application
build:
	@echo "Building application..."
	@go build -o $(BINARY_NAME) ./cmd/main.go

# Docker commands
docker-up:
	@echo "Starting Docker containers..."
	@docker compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker compose down

# Clean up
clean:
	@echo "Cleaning up..."
	@go clean
	@rm -f $(BINARY_NAME)
	@docker-compose down -v
	@kind delete cluster

# Generate Kubernetes manifests
generate-manifests:
	@echo "Generating Kubernetes manifests..."
	@mkdir -p ./deploy/manifests
	@helm template ./deploy/helm-charts/k8s-platform -f ./deploy/helm-charts/values.yaml > ./deploy/manifests/all.yaml

# Apply Kubernetes manifests to local cluster
apply-manifests:
	@echo "Applying Kubernetes manifests to local cluster..."
	@kubectl apply -f ./deploy/manifests/all.yaml

# Setup local Kubernetes development environment
k8s-setup: generate-manifests apply-manifests
	@echo "Kubernetes development environment setup complete"