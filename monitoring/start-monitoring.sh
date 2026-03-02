#!/bin/bash

set -e

echo "=== Kafka Monitoring Stack Start ==="
echo

echo "Starting all services (Kafka + Monitoring)..."
docker-compose up -d

echo
echo "Waiting for services to be ready..."
sleep 10

echo
echo "=== All services started! ==="
echo
echo "Access URLs:"
echo "  - Prometheus:     http://localhost:9090"
echo "  - Grafana:        http://localhost:3000 (admin/admin)"
echo "  - Alertmanager:   http://localhost:9093"
echo
echo "Kafka Brokers:"
echo "  - Cluster 1 (kafka-0/1/2): localhost:9094/9095/9096"
echo "  - Cluster 2 (kafka-3/4/5): localhost:9097/9098/9099"
echo
echo "JMX Exporter endpoints:"
echo "  - kafka-0-jmx: localhost:7071"
echo "  - kafka-1-jmx: localhost:7072"
echo "  - kafka-2-jmx: localhost:7073"
echo "  - kafka-3-jmx: localhost:7074"
echo "  - kafka-4-jmx: localhost:7075"
echo "  - kafka-5-jmx: localhost:7076"
