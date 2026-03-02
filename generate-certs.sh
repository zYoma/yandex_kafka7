#!/bin/bash

set -e

PASSWORD=password
VALIDITY_DAYS=365
BASE_DIR=$(pwd)/kafka-creds

mkdir -p ${BASE_DIR}

# Generate CA
echo "Generating CA..."
openssl req -new -x509 -keyout ${BASE_DIR}/ca.key -out ${BASE_DIR}/ca.crt -days ${VALIDITY_DAYS} -passout pass:${PASSWORD} \
  -subj "/C=RU/ST=State/L=City/O=Organization/CN=Kafka-CA"

# Create cluster 2 certificates (kafka-3, kafka-4, kafka-5)
CLUSTER_NODES=("kafka-0" "kafka-1" "kafka-2")

for NODE in "${CLUSTER_NODES[@]}"; do
  echo "Generating certificates for ${NODE}..."
  
  NODE_DIR=${BASE_DIR}/${NODE}-creds
  mkdir -p ${NODE_DIR}

  # Create openssl config with SAN
  cat > ${NODE_DIR}/${NODE}.cnf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
x509_extensions = v3_ca

[req_distinguished_name]
CN = ${NODE}

[v3_req]
subjectAltName = @alt_names

[v3_ca]
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${NODE}
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

  # Generate private key and CSR
  openssl genrsa -out ${NODE_DIR}/${NODE}.key 2048
  openssl req -new -key ${NODE_DIR}/${NODE}.key -out ${NODE_DIR}/${NODE}.csr \
    -subj "/C=RU/ST=State/L=City/O=Organization/CN=${NODE}" \
    -config ${NODE_DIR}/${NODE}.cnf

  # Sign with CA
  openssl x509 -req -in ${NODE_DIR}/${NODE}.csr -CA ${BASE_DIR}/ca.crt \
    -CAkey ${BASE_DIR}/ca.key -CAcreateserial -out ${NODE_DIR}/${NODE}.crt \
    -days ${VALIDITY_DAYS} -extensions v3_req -extfile ${NODE_DIR}/${NODE}.cnf \
    -passin pass:${PASSWORD}

  # Import CA into keystore
  keytool -import -alias CARoot -file ${BASE_DIR}/ca.crt \
    -keystore ${NODE_DIR}/kafka.keystore.jks -storepass ${PASSWORD} -noprompt

  # Import signed certificate
  openssl pkcs12 -export -name ${NODE} -in ${NODE_DIR}/${NODE}.crt \
    -inkey ${NODE_DIR}/${NODE}.key -out ${NODE_DIR}/${NODE}.p12 \
    -passout pass:${PASSWORD}

  keytool -importkeystore -deststorepass ${PASSWORD} -destkeypass ${PASSWORD} \
    -destkeystore ${NODE_DIR}/kafka.keystore.jks \
    -srckeystore ${NODE_DIR}/${NODE}.p12 -srcstoretype PKCS12 \
    -srcstorepass ${PASSWORD} -alias ${NODE} -noprompt

  # Create truststore
  keytool -import -alias CARoot -file ${BASE_DIR}/ca.crt \
    -keystore ${NODE_DIR}/kafka.truststore.jks -storepass ${PASSWORD} -noprompt

  # Cleanup
  rm ${NODE_DIR}/${NODE}.csr ${NODE_DIR}/${NODE}.crt ${NODE_DIR}/${NODE}.key \
     ${NODE_DIR}/${NODE}.p12 ${NODE_DIR}/${NODE}.cnf

  echo "Generated certificates for ${NODE}"
done

echo "Certificate generation complete!"
echo "Certificates stored in: ${BASE_DIR}"