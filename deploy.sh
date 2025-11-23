#!/bin/bash

# 部署 test-server 到 Kubernetes

set -e

echo "Building Docker image..."
docker build -t test-server:latest .

echo "Deploying to Kubernetes..."
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=60s deployment/test-server

echo "Deployment complete!"
echo ""
echo "To check the status:"
echo "  kubectl get pods -l app=test-server"
echo "  kubectl get svc test-server"
echo ""
echo "To view logs:"
echo "  kubectl logs -l app=test-server"
echo ""
echo "To access metrics (port-forward):"
echo "  kubectl port-forward svc/test-server 8080:8080"
echo "  curl http://localhost:8080/metrics"

