#!/bin/bash

# local-dev-setup.sh - Set up local development environment

set -e

echo "Setting up local development environment..."

# Install required tools - if the download for a dependency fails, run `go get` on the dependency and re-run the setup script
echo "Installing required tools..."
go install sigs.k8s.io/kind@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air  # For hot reloading

# Stop and remove existing Kind cluster if it exists
echo "Checking for existing Kind cluster..."
if kind get clusters | grep -q "^kind$"; then
    echo "A Kind cluster already exists. Deleting it..."
    kind delete cluster
fi

# Create kind cluster
echo "Creating Kind cluster..."
cat << EOF > kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30080
    hostPort: 30080
    protocol: TCP
- role: worker
- role: worker
EOF

kind create cluster --config kind-config.yaml

# Stop and remove existing PostgreSQL container if it exists
echo "Checking for existing PostgreSQL container..."
if docker ps -a --format '{{.Names}}' | grep -q "^postgres-k8s-platform$"; then
    echo "Stopping and removing existing PostgreSQL container..."
    docker stop postgres-k8s-platform
    docker rm postgres-k8s-platform
fi

# Set up local PostgreSQL with Docker
echo "Setting up PostgreSQL..."
docker run --name postgres-k8s-platform -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=k8s_platform -p 5432:5432 -d postgres:13-alpine

# Stop and remove existing MongoDB container if it exists
echo "Checking for existing MongoDB container..."
if docker ps -a --format '{{.Names}}' | grep -q "^mongodb-k8s-platform$"; then
    echo "Stopping and removing existing MongoDB container..."
    docker stop mongodb-k8s-platform
    docker rm mongodb-k8s-platform
fi

# Set up local MongoDB with Docker
echo "Setting up MongoDB..."
docker run --name mongodb-k8s-platform -p 27017:27017 -d mongo:5.0

# Wait for databases to be ready
echo "Waiting for databases to be ready..."
sleep 5

# Run migrations
echo "Running database migrations..."
chmod +x ./scripts/migrate.sh
./scripts/migrate.sh up

echo "Local development environment setup complete!"
echo "You can now run the application with: go run ./cmd/main.go"
