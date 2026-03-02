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

echo "=== Creating topic: ${TOPIC_NAME} ==="

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
echo "=== Setting up ACLs for topic: ${TOPIC_NAME} ==="

# For shop_client: Write only
echo "Adding WRITE ACL for shop_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /opt/bitnami/kafka/config/client.properties \
  --add \
  --allow-principal User:shop_client \
  --operation Write \
  --topic ${TOPIC_NAME}

echo "Adding DESCRIBE ACL for shop_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /opt/bitnami/kafka/config/client.properties \
  --add \
  --allow-principal User:shop_client \
  --operation Describe \
  --topic ${TOPIC_NAME}

# For analytic_client: Read only
echo "Adding READ ACL for analytic_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /opt/bitnami/kafka/config/client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Read \
  --topic ${TOPIC_NAME}

echo "Adding DESCRIBE ACL for analytic_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /opt/bitnami/kafka/config/client.properties \
  --add \
  --allow-principal User:analytic_client \
  --operation Describe \
  --topic ${TOPIC_NAME}

# For filter_client: Read only
echo "Adding READ ACL for filter_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /opt/bitnami/kafka/config/client.properties \
  --add \
  --allow-principal User:filter_client \
  --operation Read \
  --topic ${TOPIC_NAME}

echo "Adding DESCRIBE ACL for filter_client"
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /opt/bitnami/kafka/config/client.properties \
  --add \
  --allow-principal User:filter_client \
  --operation Describe \
  --topic ${TOPIC_NAME}

echo ""
echo "=== Listing ACLs for topic: ${TOPIC_NAME} ==="
docker exec -i ${KAFKA_CONTAINER} kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --command-config /opt/bitnami/kafka/config/client.properties \
  --list \
  --topic ${TOPIC_NAME}

echo ""
echo "=== Setup complete ==="