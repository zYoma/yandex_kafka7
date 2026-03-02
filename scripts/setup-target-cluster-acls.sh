#!/bin/bash

set -e

KAFKA_CONTAINER="yandex_kafka7-kafka-3-1"
CLIENT_CONFIG="/opt/bitnami/kafka/config/client.properties"

echo "=== Creating client configuration for target cluster admin ==="

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
echo "=== Setting up internal Kafka ACLs for target cluster admin ==="

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-3:9091 \
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
  --bootstrap-server kafka-3:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:admin \
  --operation All \
  --cluster

echo ""
echo "=== Setting up ACLs for inter-broker user on target cluster ==="

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-3:9091 \
  --command-config ${CLIENT_CONFIG} \
  --add \
  --allow-principal User:user \
  --operation ClusterAction \
  --cluster

docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-3:9091 \
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
echo "=== Listing target cluster ACLs ==="

echo "--- Cluster-level ACLs ---"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-3:9091 \
  --command-config ${CLIENT_CONFIG} \
  --list \
  --cluster

echo ""
echo "=== Target cluster setup complete ==="