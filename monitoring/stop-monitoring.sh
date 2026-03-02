#!/bin/bash

set -e

echo "=== Kafka Monitoring Stack Stop ==="
echo

echo "Stopping all services..."
cd ..
docker-compose down

echo
echo "=== All services stopped ==="
