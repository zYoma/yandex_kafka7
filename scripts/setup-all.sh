#!/bin/bash

TOPIC_NAME="shops_data"
PARTITIONS=3
REPLICATION_FACTOR=3
KAFKA_IMAGE="bitnamilegacy/kafka:3.4"
KAFKA_NETWORK="yandex_kafka7_kafka-network"
CLIENT_CONFIG_PATH=$(pwd)/kafka-creds/client.properties
CLUSTER1_CERTS_PATH=$(pwd)/kafka-creds/kafka-0-creds
CLUSTER2_CERTS_PATH=$(pwd)/kafka-creds/kafka-3-creds

run_kafka_cmd_cluster1() {
  local cmd="$@"
  docker run --rm --network ${KAFKA_NETWORK} \
    -v ${CLIENT_CONFIG_PATH}:/client.properties \
    -v ${CLUSTER1_CERTS_PATH}:/opt/bitnami/kafka/config/certs \
    -e KAFKA_OPTS='' \
    ${KAFKA_IMAGE} \
    $cmd 2>&1 | grep -v "JMX\|jdk\|rmi\|kafka \|\[38m\|\[0m" | grep -v "Welcome to\|Subscribe to\|Upgrade to\|Picked up" || true
}

run_kafka_cmd_cluster2() {
  local cmd="$@"
  docker run --rm --network ${KAFKA_NETWORK} \
    -v ${CLIENT_CONFIG_PATH}:/client.properties \
    -v ${CLUSTER2_CERTS_PATH}:/opt/bitnami/kafka/config/certs \
    -e KAFKA_OPTS='' \
    ${KAFKA_IMAGE} \
    $cmd 2>&1 | grep -v "JMX\|jdk\|rmi\|kafka \|\[38m\|\[0m" | grep -v "Welcome to\|Subscribe to\|Upgrade to\|Picked up" || true
}

echo "=== Verifying Kafka connection ==="
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-broker-api-versions.sh --bootstrap-server kafka-0:9091 --command-config /client.properties 2>&1 | head -3
echo ""

echo ""
echo "=== Step 1: Setting up internal Kafka ACLs for admin ==="

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:admin \
  --operation Read \
  --operation Write \
  --operation Describe \
  --operation DescribeConfigs \
  --topic '*' \
  --resource-pattern-type Literal

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:admin \
  --operation All \
  --cluster

echo ""
echo "=== Step 2: Setting up ACLs for Schema Registry ==="

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:admin \
  --operation Create \
  --operation Write \
  --operation Read \
  --operation Describe \
  --topic _schemas \
  --resource-pattern-type Literal

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:user \
  --operation Read \
  --operation Describe \
  --operation Alter \
  --operation AlterConfigs \
  --topic _schemas \
  --resource-pattern-type Literal

echo ""
echo "=== Step 3: Setting up ACLs for inter-broker user ==="

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:user \
  --operation ClusterAction \
  --cluster

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:user \
  --operation Read \
  --operation Write \
  --operation Describe \
  --operation DescribeConfigs \
  --operation Create \
  --operation Delete \
  --operation Alter \
  --operation AlterConfigs \
  --topic '*' \
  --resource-pattern-type Literal

echo ""
echo "=== Step 4: Creating topic: ${TOPIC_NAME} ==="

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-topics.sh \
  --create \
  --topic ${TOPIC_NAME} \
  --bootstrap-server kafka-0:9091 \
  --partitions ${PARTITIONS} \
  --replication-factor ${REPLICATION_FACTOR} \
  --command-config /client.properties \
  --if-not-exists

echo "Topic ${TOPIC_NAME} created successfully"

echo ""
echo "=== Step 5: Setting up ACLs for topic: ${TOPIC_NAME} ==="

echo "Adding WRITE ACL for shop_client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:shop_client \
  --operation Write \
  --operation Describe \
  --operation Create \
  --topic ${TOPIC_NAME}

echo "Adding READ ACL for analytic_client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Read \
  --operation Describe \
  --topic ${TOPIC_NAME}

echo "Adding READ ACL for filter_client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:filter_client \
  --operation Read \
  --operation Describe \
  --topic ${TOPIC_NAME}

echo "Adding consumer group ACL for analytic_client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Read \
  --group '*'

echo "Adding consumer group ACL for filter_client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:filter_client \
  --operation Read \
  --group '*'

echo ""
echo "=== Step 6: Creating topic: recommendations ==="

RECOMMENDATIONS_TOPIC="recommendations"

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-topics.sh \
  --create \
  --topic ${RECOMMENDATIONS_TOPIC} \
  --bootstrap-server kafka-0:9091 \
  --partitions ${PARTITIONS} \
  --replication-factor ${REPLICATION_FACTOR} \
  --command-config /client.properties \
  --if-not-exists

echo "Topic ${RECOMMENDATIONS_TOPIC} created successfully"

echo "Setting up ACLs for topic: ${RECOMMENDATIONS_TOPIC}"

echo "Adding WRITE and CREATE ACL for analytic_client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Write \
  --operation Describe \
  --operation Create \
  --topic ${RECOMMENDATIONS_TOPIC}

echo "Adding READ ACL for client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:client \
  --operation Read \
  --operation Describe \
  --topic ${RECOMMENDATIONS_TOPIC}

echo "Adding consumer group ACL for client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:client \
  --operation Read \
  --group '*'

echo ""
echo "=== Step 7: Creating topic: requests ==="

REQUESTS_TOPIC="requests"

run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-topics.sh \
  --create \
  --topic ${REQUESTS_TOPIC} \
  --bootstrap-server kafka-0:9091 \
  --partitions ${PARTITIONS} \
  --replication-factor ${REPLICATION_FACTOR} \
  --command-config /client.properties \
  --if-not-exists

echo "Topic ${REQUESTS_TOPIC} created successfully"

echo "Setting up ACLs for topic: ${REQUESTS_TOPIC}"

echo "Adding WRITE and CREATE ACL for client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:client \
  --operation Write \
  --operation Describe \
  --operation Create \
  --topic ${REQUESTS_TOPIC}

echo "Adding READ ACL for analytic_client"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Read \
  --operation Describe \
  --topic ${REQUESTS_TOPIC}

echo ""
echo "=== Step 8: Listing all ACLs ==="

echo "--- Cluster-level ACLs ---"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --list \
  --cluster

echo ""
echo "--- Wildcard topic ACLs ---"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --list \
  --topic '*'

echo ""
echo "--- _schemas topic ACLs ---"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --list \
  --topic _schemas

echo ""
echo "--- ${TOPIC_NAME} topic ACLs ---"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --list \
  --topic ${TOPIC_NAME}

echo ""
echo "--- ${RECOMMENDATIONS_TOPIC} topic ACLs ---"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --list \
  --topic ${RECOMMENDATIONS_TOPIC}

echo ""
echo "--- ${REQUESTS_TOPIC} topic ACLs ---"
run_kafka_cmd_cluster1 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /client.properties \
  --list \
  --topic ${REQUESTS_TOPIC}

echo ""
echo "========================================"
echo "=== Setting up second Kafka cluster (kafka-3:9094) ==="
echo "========================================"

echo ""
echo "=== Step 9: Setting up ACLs for admin (cluster 2) ==="
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:admin \
  --operation All \
  --cluster

run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:admin \
  --operation Read \
  --operation Write \
  --operation Describe \
  --operation DescribeConfigs \
  --topic '*' \
  --resource-pattern-type Literal

echo ""
echo "=== Step 10: Setting up ACLs for inter-broker user (cluster 2) ==="
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:user \
  --operation ClusterAction \
  --cluster

run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:user \
  --operation Read \
  --operation Write \
  --operation Describe \
  --operation DescribeConfigs \
  --operation Create \
  --operation Delete \
  --operation Alter \
  --operation AlterConfigs \
  --topic '*' \
  --resource-pattern-type Literal

echo ""
echo "=== Step 11: Setting up ACLs for Schema Registry topic (cluster 2) ==="
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:user \
  --operation Read \
  --operation Describe \
  --operation Alter \
  --operation AlterConfigs \
  --topic _schemas \
  --resource-pattern-type Literal

echo ""
echo "=== Step 11.5: Setting up ACLs for analytic_client (cluster 2) ==="

echo "Adding READ ACL for analytic_client on requests topic"
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Read \
  --operation Describe \
  --topic requests

echo "Adding WRITE and CREATE ACL for analytic_client on recommendations topic"
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Write \
  --operation Describe \
  --operation Create \
  --topic recommendations

echo "Adding consumer group ACL for analytic_client"
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Read \
  --group '*'

echo ""
echo "=== Step 12: Listing Cluster 2 ACLs ==="
echo "--- Cluster-level ACLs (cluster 2) ---"
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --list \
  --cluster

echo ""
echo "--- Wildcard topic ACLs (cluster 2) ---"
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --list \
  --topic '*'

echo ""
echo "--- _schemas topic ACLs (cluster 2) ---"
run_kafka_cmd_cluster2 /opt/bitnami/kafka/bin/kafka-acls.sh \
  --bootstrap-server kafka-3:9094 \
  --command-config /client.properties \
  --list \
  --topic _schemas

echo ""
echo "=== All setup complete ==="
