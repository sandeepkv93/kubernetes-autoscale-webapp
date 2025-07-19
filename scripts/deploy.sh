#!/bin/bash
set -e

# Add Docker to PATH if needed
export PATH="/Applications/Docker.app/Contents/Resources/bin:$PATH"

echo "Building Docker images..."
# Build backend
cd backend
docker build -t backend:latest .
cd ..

# Build frontend
cd frontend
docker build -t frontend:latest .
cd ..

echo "Loading images into kind cluster..."
kind load docker-image backend:latest --name webapp-cluster
kind load docker-image frontend:latest --name webapp-cluster

echo "Deploying to Kubernetes..."
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Deploy metrics server for HPA
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Wait for metrics server
echo "Waiting for metrics-server to be ready..."
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=60s

# Deploy configurations
kubectl apply -f k8s/secrets/
kubectl apply -f k8s/configmaps/

# Deploy database
kubectl apply -f k8s/database/

# Wait for database to be ready
echo "Waiting for database to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n webapp --timeout=120s

# Deploy Redis
kubectl apply -f k8s/redis/

# Deploy backend
kubectl apply -f k8s/backend/

# Deploy frontend
kubectl apply -f k8s/frontend/

# Deploy ingress
kubectl apply -f k8s/ingress/

echo "Deployment completed!"
echo "Add '127.0.0.1 webapp.local' to your /etc/hosts file"
echo "Access the application at http://webapp.local"