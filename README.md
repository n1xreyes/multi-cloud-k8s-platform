# Multi-Cloud Kubernetes Application Deployment Platform

This project is an attempt at creating a platform that allows developers to deploy applications across multiple cloud providers through a unified interface.

## Running the project locally (only for the database for now):

On the command terminal at the project root
1. Setup the local environment:  
```make setup```
2. Start all required services using Docker Compose:  
```make docker-up```
3. Start up the local K8S cluster
```make local-dev-setup```
4. Run database migrations:  
```make migrate-up```

Refer to the **_[makefile](https://github.com/n1xreyes/multi-cloud-k8s-platform/blob/main/makefile)_** for more targets to manage the application. 