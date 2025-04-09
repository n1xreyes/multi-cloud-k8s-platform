#!/bin/bash

# setup.sh - Verify and setup development environment for Multi-Cloud K8s Platform

# Check for required tools
echo "Checking prerequisites..."

# Check for Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "✓ Go found (version $GO_VERSION)"
    # Check minimum version (1.21)
    if [[ $(echo "$GO_VERSION" | awk -F. '{print $1*100+$2}') -lt 121 ]]; then
        echo "⚠ Go version should be at least 1.21. Please update."
    fi
else
    echo "✗ Go not found. Please install Go 1.21 or later."
    echo "  Visit https://golang.org/doc/install for installation instructions."
fi

# Check for Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    echo "✓ Docker found (version $DOCKER_VERSION)"
else
    echo "✗ Docker not found. Please install Docker."
    echo "  Visit https://docs.docker.com/get-docker/ for installation instructions."
fi

# Check for kubectl
if command -v kubectl &> /dev/null; then
    KUBECTL_VERSION=$(kubectl version --client --short | awk '{print $3}')
    echo "✓ kubectl found (version $KUBECTL_VERSION)"
else
    echo "✗ kubectl not found. Please install kubectl."
    echo "  Visit https://kubernetes.io/docs/tasks/tools/install-kubectl/ for installation instructions."
fi

# Check for minikube
if command -v minikube &> /dev/null; then
    MINIKUBE_VERSION=$(minikube version | awk '{print $3}')
    echo "✓ minikube found (version $MINIKUBE_VERSION)"
else
    echo "✗ minikube not found. Please install minikube."
    echo "  Visit https://minikube.sigs.k8s.io/docs/start/ for installation instructions."
fi

# Check if minikube is running
if command -v minikube &> /dev/null; then
    MINIKUBE_STATUS=$(minikube status -f {{.Host}} 2>/dev/null)
    if [[ "$MINIKUBE_STATUS" == "Running" ]]; then
        echo "✓ minikube is running"
    else
        echo "⚠ minikube is not running. Starting minikube..."
        minikube start
    fi
fi

echo -e "\nSetting up project dependencies..."
# Initialize Go modules and get initial dependencies
go mod tidy
go get -u k8s.io/client-go@latest
go get -u sigs.k8s.io/controller-runtime@latest
go get -u k8s.io/apimachinery@latest
go get -u k8s.io/api@latest
go get -u sigs.k8s.io/cluster-api@latest
go get -u github.com/gorilla/mux@latest
go get -u go.mongodb.org/mongo-driver@latest
go get -u github.com/lib/pq@latest
go get -u github.com/spf13/cobra@latest
go get -u github.com/spf13/viper@latest

echo -e "\nCreating git repository..."
if [ ! -d .git ]; then
    git init
    echo "# Multi-Cloud Kubernetes Application Deployment Platform" > README.md
    echo "/bin/node_modules" > .gitignore
    echo "*.swp" >> .gitignore
    echo ".env" >> .gitignore
    echo ".idea/" >> .gitignore
    echo ".vscode/" >> .gitignore
    echo "*.exe" >> .gitignore
    git add .
    git commit -m "Initial project setup"
fi


# Replace <YOUR_GITHUB_USERNAME> and <YOUR_REPOSITORY_NAME> with actual values
#git remote add origin https://github.com/<YOUR_GITHUB_USERNAME>/<YOUR_REPOSITORY_NAME>.git
#git branch -M main
#git push -u origin main

echo -e "\nSetup complete! Development environment is ready."
echo "Next steps:"
echo "1. Define Custom Resource Definitions (CRDs)"
echo "2. Implement the Kubernetes operator"
echo "3. Create the REST API service"
echo "4. Develop the microservices"