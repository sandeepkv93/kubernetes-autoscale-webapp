#!/bin/bash

echo "Cleaning up resources..."
kubectl delete namespace webapp
echo "Cleanup completed!"