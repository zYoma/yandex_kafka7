#!/bin/bash

# Скрипт для запуска JMX Exporter с подключением к JMX порту Kafka

set -e

JMX_HOST=${JMX_HOST:-localhost}
JMX_PORT=${JMX_PORT:-9999}
CONFIG_FILE=${JMX_EXPORTER_CONFIG_FILE:-/opt/jmx-exporter/config.yml}
EXPORTER_PORT=${JMX_EXPORTER_PORT:-7071}

echo "Starting JMX Exporter..."
echo "JMX Host: $JMX_HOST"
echo "JMX Port: $JMX_PORT"
echo "Config: $CONFIG_FILE"
echo "Exporter Port: $EXPORTER_PORT"

# Создаем временный конфиг с serviceUrl
TEMP_CONFIG="/tmp/jmx_exporter_config.yml"
cat "$CONFIG_FILE" > "$TEMP_CONFIG"
echo "" >> "$TEMP_CONFIG"
echo "jmxUrl: \"service:jmx:rmi:///jndi/rmi://$JMX_HOST:$JMX_PORT/jmxrmi\"" >> "$TEMP_CONFIG"

# Запускаем JMX Exporter HTTP server для чтения JMX метрик через JMX соединение
java -jar /opt/jmx-exporter/jmx_prometheus_httpserver.jar "$EXPORTER_PORT" "$TEMP_CONFIG"
