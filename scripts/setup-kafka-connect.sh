#!/bin/bash

set -e

echo "Setting up Kafka Connect ACLs and topics..."
echo ""

CONTAINER_NAME="yandex_kafka7-kafka-0-1"

echo "Waiting for Kafka Connect to be ready..."
sleep 5

echo ""
echo "Creating Kafka Connect topics..."
docker exec "$CONTAINER_NAME" kafka-topics.sh \
  --bootstrap-server kafka-0:9091 \
  --create \
  --topic connect-configs \
  --partitions 1 \
  --replication-factor 3 \
  --config cleanup.policy=compact \
  --command-config /opt/bitnami/kafka/config/client.properties || echo "Topic connect-configs may already exist"

docker exec "$CONTAINER_NAME" kafka-topics.sh \
  --bootstrap-server kafka-0:9091 \
  --create \
  --topic connect-offsets \
  --partitions 25 \
  --replication-factor 3 \
  --config cleanup.policy=compact \
  --command-config /opt/bitnami/kafka/config/client.properties || echo "Topic connect-offsets may already exist"

docker exec "$CONTAINER_NAME" kafka-topics.sh \
  --bootstrap-server kafka-0:9091 \
  --create \
  --topic connect-status \
  --partitions 5 \
  --replication-factor 3 \
  --config cleanup.policy=compact \
  --command-config /opt/bitnami/kafka/config/client.properties || echo "Topic connect-status may already exist"

echo ""
echo "Setting up ACLs for filter_client..."
echo ""

docker exec "$CONTAINER_NAME" kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --add \
  --allow-principal User:filter_client \
  --operation READ \
  --topic shops_data_json \
  --command-config /opt/bitnami/kafka/config/client.properties

docker exec "$CONTAINER_NAME" kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --add \
  --allow-principal User:filter_client \
  --operation DESCRIBE \
  --topic shops_data_json \
  --command-config /opt/bitnami/kafka/config/client.properties

docker exec "$CONTAINER_NAME" kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --add \
  --allow-principal User:filter_client \
  --operation ALL \
  --group connect-\* \
  --resource-pattern-type PREFIXED \
  --force \
  --command-config /opt/bitnami/kafka/config/client.properties

echo ""
echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Build the plugin: cd kafka-connect && ./build-plugin.sh"
echo "2. Restart Kafka Connect: docker compose restart kafka-connect"
echo "3. Deploy the connector: ../scripts/connector-manager.sh deploy"
