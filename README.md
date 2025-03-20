# Multi-Cloud Kubernetes Application Deployment Platform

This project is an attempt at creating a platform that allows developers to deploy applications across multiple cloud providers through a unified interface.

## Running the project locally (only for the database for now):

On the command terminal at the project root
1. Setup the local environent:  
```make setup```
2. Start all required services using Docker Compose:  
```make docker-up```
3. Run database migrations:  
```make migrate-up```
4. Run the application in development mode:
```make dev```

## Local testing
1. Run all tests including integration tests
```make test```
2. Run only unit tests (skipping integration tests)
```make test-short```
3. Test local Kubernetes deployment:
```make k8s-setup```
4. Clean up after testing:
```make clean```