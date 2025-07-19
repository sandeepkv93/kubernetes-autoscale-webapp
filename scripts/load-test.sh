#!/bin/bash

echo "Starting load test..."
echo "This will generate CPU load to trigger HPA"

# Get the backend service endpoint
BACKEND_URL="http://webapp.local/api/stress"

# Run concurrent requests
echo "Sending 100 concurrent requests..."
for i in {1..100}; do
  curl -s $BACKEND_URL &
done

wait

echo "Load test completed!"
echo "Check HPA status with: kubectl get hpa -n webapp -w"