#!/bin/bash

echo "Starting intensive load test..."

# Function to continuously send requests with no delay
send_requests() {
  while true; do
    curl -s http://localhost:8080/api/stress > /dev/null
  done
}

# Start multiple background processes with higher concurrency
for i in {1..50}; do
  send_requests &
done

echo "Intensive load test running with 50 concurrent processes..."
echo "Press Ctrl+C to stop"
wait