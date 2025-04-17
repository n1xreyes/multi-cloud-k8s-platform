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
	go mod tidy

# Build all services
build:
	$(GO) build -o ./bin/api-server ./cmd/api-server/main.go
	$(GO) build -o ./bin/api-gateway ./cmd/api-gateway/main.go
	$(GO) build -o ./bin/config-server ./cmd/config-server/main.go
	# Uncomment when operator and cli exist and are ready
	#$(GO) build -o ./bin/operator ./cmd/operator/main.go
	#$(GO) build -o ./bin/cli ./cmd/cli/main.go

# Run specific service
run-api:
	$(GO) run ./cmd/api-server/main.go

run-gateway:
	$(GO) run ./cmd/api-gateway/main.go

run-config:
	$(GO) run ./cmd/config-server/main.go

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
get_pg_pod_name = $(shell kubectl get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

migrate-up:
	@echo "Waiting for Postgres pod to be ready..."
	@kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s
	@POD_NAME=$$(kubectl get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	 if [ -z "$$POD_NAME" ]; then echo "ERROR: Postgres pod not found."; exit 1; fi; \
	 echo "Executing migrations in pod $$POD_NAME..."; \
	 kubectl exec "$$POD_NAME" -- /app/scripts/migrate.sh up

migrate-down:
	@echo "Waiting for Postgres pod to be ready..."
	@kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s
	@POD_NAME=$$(kubectl get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	 if [ -z "$$POD_NAME" ]; then echo "ERROR: Postgres pod not found."; exit 1; fi; \
	 echo "Executing migration down in pod $$POD_NAME..."; \
	 kubectl exec "$$POD_NAME" -- /app/scripts/migrate.sh down

migrate-reset:
	@echo "Waiting for Postgres pod to be ready..."
	@kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s
	@POD_NAME=$$(kubectl get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	 if [ -z "$$POD_NAME" ]; then echo "ERROR: Postgres pod not found."; exit 1; fi; \
	 echo "Executing migration reset in pod $$POD_NAME..."; \
	 kubectl exec "$$POD_NAME" -- /app/scripts/migrate.sh reset

migrate-create:
	./scripts/migrate.sh create $(name)

# Docker operations
docker-build:
	# Create build directories for all services defined in docker-compose
	mkdir -p build/api build/auth build/config build/gateway
	cp build/api/Dockerfile build/api/Dockerfile
	cp build/auth/Dockerfile build/auth/Dockerfile
	cp build/config/Dockerfile build/config/Dockerfile
	cp build/gateway/Dockerfile build/gateway/Dockerfile
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
	fi
	@echo "Ensuring addons are enabled..."
	@$(MINIKUBE) addons enable ingress
	@$(MINIKUBE) addons enable dashboard
	@$(MINIKUBE) addons enable metrics-server
	@echo "Minikube cluster is up and core addons enabled."
	@echo "Waiting a few seconds for addons to stabilize..."
	@sleep 5 # Give dashboard time to potentially start

cluster-down:
	$(MINIKUBE) stop

cluster-delete:
	$(MINIKUBE) delete

# Local Kubernetes deployment
deploy-local: cluster-up
	@echo "Creating build directories if they don't exist..."
	@mkdir -p build/api build/auth build/config build/gateway
	@echo "Copying Dockerfiles..."
	@cp -n build/auth/Dockerfile build/auth/Dockerfile 2>/dev/null || echo "Skipping auth.Dockerfile copy"
	@cp -n build/api/Dockerfile build/api/Dockerfile 2>/dev/null || echo "Skipping api.Dockerfile copy"
	@cp -n build/config/Dockerfile build/config/Dockerfile 2>/dev/null || echo "Skipping config.Dockerfile copy"
	@cp -n build/gateway/Dockerfile build/gateway/Dockerfile 2>/dev/null || echo "Skipping gateway.Dockerfile copy"
	@cp -n build/postgres/Dockerfile build/postgres/Dockerfile 2>/dev/null || echo "Skipping postgres.Dockerfile copy"
	@echo "Setting docker environment to use Minikube's Docker daemon..."
	@eval $$(minikube -p minikube docker-env) ;\
	echo "Building images within Minikube's Docker daemon..."; \
	$(DOCKER) build -t multi-cloud-k8s/api-server:dev -f build/api/Dockerfile . ; \
	$(DOCKER) build -t multi-cloud-k8s/auth-server:dev -f build/auth/Dockerfile . ; \
	$(DOCKER) build -t multi-cloud-k8s/config-server:dev -f build/config/Dockerfile . ; \
	$(DOCKER) build -t multi-cloud-k8s/postgres-custom:dev -f build/postgres/Dockerfile . ; \
	$(DOCKER) build -t multi-cloud-k8s/gateway-server:dev -f build/gateway/Dockerfile .
	@echo "Deploying services to Minikube..."
	$(KUBECTL) apply -f deployments/local/

undeploy-local:
	$(KUBECTL) delete -f deployments/local/

# Set up local development environment
local-dev-setup: setup cluster-up deploy-local migrate-up
	@echo "Local development environment is set up and running"
	@echo "Ensure DB migrations are applied: make migrate-up"
	@echo "---"
	@echo "Access via NodePorts (if applicable):"
	@echo "API Server (NodePort): http://$$(minikube ip):$$(kubectl get svc api -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo N/A)"
	@echo "Auth Server (NodePort): http://$$(minikube ip):$$(kubectl get svc auth -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo N/A)"
	@echo "Config Server (NodePort): http://$$(minikube ip):$$(kubectl get svc config-service -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo N/A)"
	@echo "___"
	@echo "Access via Gateway (if deployed with Ingress/NodePort):"
	@# Adjust gateway service name if different
	@GATEWAY_NODEPORT=$$(kubectl get svc gateway-service -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null)
	@if [ -n "$GATEWAY_NODEPORT" ]; then \
		GATEWAY_IP=$$(minikube ip); \
		echo "Gateway Base URL: http://$${GATEWAY_IP}:$${GATEWAY_NODEPORT}/api/v1"; \
		echo "Example Config Endpoint: http://$${GATEWAY_IP}:$${GATEWAY_NODEPORT}/api/v1/configs"; \
	else \
		echo "Gateway (NodePort): N/A"; \
	fi
	@echo "---"
	@echo "Attempting to get Minikube Dashboard URL..."
	@# Try to get the dashboard URL. Suppress proxy messages (stderr) and provide fallback instructions.
	@# This might still take a moment or fail if the proxy isn't ready, but shouldn't hang the whole build.
	@DASHBOARD_URL=$$(timeout 15s minikube dashboard --url 2>/dev/null) || DASHBOARD_URL=""
	@if [ -n "$DASHBOARD_URL" ]; then \
		echo "Dashboard URL: $${DASHBOARD_URL}"; \
	else \
		echo "Could not retrieve dashboard URL automatically. Please run 'minikube dashboard' in a separate terminal."; \
	fi

# Development convenience targets
dev: docker-up
	@echo "Development environment is up and running"

dev-reset: docker-down clean docker-build docker-up migrate-reset
	@echo "Development environment has been reset"