#!/bin/bash

echo "Starting sustained load test..."

# Function to continuously send requests
send_requests() {
  while true; do
    curl -s http://localhost:8080/api/stress > /dev/null
    sleep 0.1
  done
}

# Start multiple background processes
for i in {1..20}; do
  send_requests &
done

echo "Load test running with 20 concurrent processes..."
echo "Press Ctrl+C to stop"
wait