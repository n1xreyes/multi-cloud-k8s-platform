# Multi-Cloud Kubernetes Application Deployment Platform

This project is an attempt at creating a platform that allows developers to deploy applications across multiple cloud providers through a unified interface.

## To run locally using Kubernetes (Minikube):

On the command terminal at the project root
1. Ensure Prerequisites: Make sure go, docker, kubectl, and minikube are installed and working.
2. Tidy Go modules:  
```go mod tidy```
3. Run Setup & Deployment
```make local-dev-setup```  
   This will:  
    - Run ```setup.sh```
    - Ensure Minikube is running (cluster-up).
    - Copy Dockerfiles (if needed).
    - Build the Go binaries (including config-server).
    - Build Docker images within Minikube's Docker environment (including config-server:dev).
    - Apply all Kubernetes manifests in deployments/local/ (including config.yaml).
    - Run database migrations (migrate-up).
4. Verify: Check the output of make local-dev-setup for the NodePort URL of the config-service (e.g., http://<```minikube_ip```>:30082).
5. Interact: You can now send requests to the Configuration Service via its NodePort or through the API Gateway (if you deployed and configured it) at `````/api/v1/configs`````. Remember to include the ```Authorization: Bearer <token>``` header if going via the gateway.

```aiignore
# Create config (replace user ID if needed)
curl -X POST http://$(minikube ip):30082/configs \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \ # Add user ID header
  -d '{"name": "my-app-config", "namespace": "dev", "configData": {"url": "http://example.com", "retries": 3}}'

# Get config
curl http://$(minikube ip):30082/configs/my-app-config?namespace=dev \
    -H "X-User-ID: 1" # Add user ID header

# List configs
curl http://$(minikube ip):30082/configs?namespace=dev \
    -H "X-User-ID: 1" # Add user ID header
```

Refer to the **_[makefile](https://github.com/n1xreyes/multi-cloud-k8s-platform/blob/main/makefile)_** for more targets to manage the application. 