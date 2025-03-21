.PHONY: setup build run clean test migrate-up migrate-down docker-up docker-down cluster-up cluster-down cluster-status deploy-local

# Default to development environment
ENV ?= development

# Variables
GO=go
DOCKER=docker
DOCKER_COMPOSE=docker compose
KUBECTL=kubectl
MINIKUBE=minikube

# Setup development environment
setup:
	./scripts/setup.sh

# Build all services
build:
	$(GO) build -o ./bin/api-server ./cmd/api-server/main.go
	$(GO) build -o ./bin/api-gateway ./cmd/api-gateway/main.go
	$(GO) build -o ./bin/operator ./cmd/operator/main.go
	$(GO) build -o ./bin/cli ./cmd/cli/main.go

# Run specific service
run-api:
	$(GO) run ./cmd/api-server/main.go

run-gateway:
	$(GO) run ./cmd/api-gateway/main.go

run-operator:
	$(GO) run ./cmd/operator/main.go

run-cli:
	$(GO) run ./cmd/cli/main.go

# Clean build artifacts
clean:
	rm -rf ./bin/*

# Run tests
test:
	$(GO) test ./...

# Database migrations
migrate-up:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down

migrate-reset:
	./scripts/migrate.sh reset

migrate-create:
	./scripts/migrate.sh create $(name)

# Docker operations
docker-build:
	mkdir -p build/api build/auth
	cp dockerfiles/api.Dockerfile build/api/Dockerfile
	cp dockerfiles/auth.Dockerfile build/auth/Dockerfile
	$(DOCKER_COMPOSE) build

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

# Kubernetes cluster management
cluster-status:
	$(MINIKUBE) status

cluster-up:
	@echo "Checking Minikube status..."
	@if $(MINIKUBE) status -f '{{.Host}}' 2>/dev/null | grep -q "Running"; then \
		echo "Minikube is already running."; \
	else \
		echo "Starting Minikube..."; \
		$(MINIKUBE) start --memory=4096 --cpus=2; \
		echo "Enabling addons..."; \
		$(MINIKUBE) addons enable ingress; \
		$(MINIKUBE) addons enable dashboard; \
		$(MINIKUBE) addons enable metrics-server; \
	fi
	@echo "Kubernetes cluster is running."
	@echo "Dashboard available at: $$(minikube dashboard --url)"

cluster-down:
	$(MINIKUBE) stop

cluster-delete:
	$(MINIKUBE) delete

# Local Kubernetes deployment
deploy-local: docker-build
	@echo "Setting docker environment to use Minikube's Docker daemon..."
	@eval $$(minikube -p minikube docker-env) && \
	$(DOCKER) build -t multi-cloud-k8s/api-server:dev -f build/api/Dockerfile . && \
	$(DOCKER) build -t multi-cloud-k8s/auth-server:dev -f build/auth/Dockerfile .
	@echo "Deploying services to Minikube..."
	$(KUBECTL) apply -f deployments/local/

undeploy-local:
	$(KUBECTL) delete -f deployments/local/

# Set up local development environment
local-dev-setup: setup docker-build cluster-up deploy-local
	@echo "Local development environment is set up and running"
	@echo "API server available at: http://$$(minikube ip):$$(kubectl get svc api -o jsonpath='{.spec.ports[0].nodePort}')"
	@echo "Auth server available at: http://$$(minikube ip):$$(kubectl get svc auth -o jsonpath='{.spec.ports[0].nodePort}')"

# Development convenience targets
dev: docker-up
	@echo "Development environment is up and running"

dev-reset: docker-down clean docker-up migrate-reset
	@echo "Development environment has been reset"