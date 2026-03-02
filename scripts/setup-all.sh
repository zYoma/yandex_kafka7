#!/bin/bash

set -e

KAFKA_CONTAINER="yandex_kafka7-kafka-0-1"
TOPIC_NAME="shops_data"
PARTITIONS=3
REPLICATION_FACTOR=3
CLIENT_CONFIG="/opt/bitnami/kafka/config/client.properties"

echo "=== Creating client configuration file ==="

docker exec -i ${KAFKA_CONTAINER} bash -c "cat > ${CLIENT_CONFIG} << 'EOF'
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username=\"admin\" password=\"admin-secret\";
ssl.truststore.location=/opt/bitnami/kafka/config/certs/kafka.truststore.jks
ssl.truststore.password=password
ssl.endpoint.identification.algorithm=
EOF
"

echo ""
echo "=== Step 1: Setting up internal Kafka ACLs for admin ==="

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:admin \
  --operation Read \
  --operation Write \
  --operation Describe \
  --operation DescribeConfigs \
  --topic '*' \
  --resource-pattern-type Literal

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:admin \
  --operation All \
  --cluster

echo ""
echo "=== Step 2: Setting up ACLs for Schema Registry ==="

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:admin \
  --operation Create \
  --operation Write \
  --operation Read \
  --operation Describe \
  --topic _schemas \
  --resource-pattern-type Literal

echo ""
echo "=== Step 3: Setting up ACLs for inter-broker user ==="

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:user \
  --operation ClusterAction \
  --cluster

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:user \
  --operation Read \
  --operation Write \
  --operation Describe \
  --operation DescribeConfigs \
  --operation Create \
  --topic '*' \
  --resource-pattern-type Literal

echo ""
echo "=== Step 4: Creating topic: ${TOPIC_NAME} ==="

docker exec -i ${KAFKA_CONTAINER} kafka-topics.sh \
  --create \
  --topic ${TOPIC_NAME} \
  --bootstrap-server kafka-0:9091 \
  --partitions ${PARTITIONS} \
  --replication-factor ${REPLICATION_FACTOR} \
  --command-config ${CLIENT_CONFIG} \
  --if-not-exists

echo "Topic ${TOPIC_NAME} created successfully"

echo ""
echo "=== Step 5: Setting up ACLs for topic: ${TOPIC_NAME} ==="

echo "Adding WRITE ACL for shop_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:shop_client \
  --operation Write \
  --operation Describe \
  --topic ${TOPIC_NAME}

echo "Adding READ ACL for analytic_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:analytic_client \
  --operation Read \
  --operation Describe \
  --topic ${TOPIC_NAME}

echo "Adding READ ACL for filter_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:filter_client \
  --operation Read \
  --operation Describe \
  --topic ${TOPIC_NAME}

echo ""
echo "=== Step 6: Listing all ACLs ==="

echo "--- Cluster-level ACLs ---"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --list \
  --cluster

echo ""
echo "--- Wildcard topic ACLs ---"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --list \
  --topic '*'

echo ""
echo "--- _schemas topic ACLs ---"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --list \
  --topic _schemas

echo ""
echo "--- ${TOPIC_NAME} topic ACLs ---"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config ${CLIENT_CONFIG} \
  --list \
  --topic ${TOPIC_NAME}

echo ""
echo "=== All setup complete ==="